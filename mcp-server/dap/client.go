package dap

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

type Client struct {
	logger      *slog.Logger
	session     *Session
	process     *exec.Cmd
	lastProject string
	OnOutput    func(category, text string)
}

func NewClient(logger *slog.Logger) *Client {
	return &Client{logger: logger}
}

func (c *Client) Launch(projectPath string, args []string) error {
	if c.session != nil {
		return fmt.Errorf("session already active, call Disconnect first")
	}

	c.lastProject = projectPath

	dll, err := c.buildProject(projectPath)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	c.process = exec.Command("netcoredbg", "--interpreter=vscode")
	c.process.Stderr = os.Stderr

	stdin, err := c.process.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := c.process.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	if err := c.process.Start(); err != nil {
		return fmt.Errorf("start netcoredbg: %w", err)
	}

	c.logger.Info("netcoredbg started", "pid", c.process.Process.Pid)

	transport := NewTransport(stdout, stdin, c.logger)
	c.session = NewSession(transport, c.logger)
	c.session.OnOutput = c.OnOutput
	c.session.StartReceiveLoop()

	if err := c.session.Initialize(); err != nil {
		c.cleanup()
		return err
	}

	if err := c.session.SetExceptionBreakpoints([]string{"unhandled"}); err != nil {
		c.logger.Warn("could not set exception breakpoints", "error", err)
	}

	if err := c.session.Launch(dll, args); err != nil {
		c.cleanup()
		return err
	}

	if err := c.session.ConfigurationDone(); err != nil {
		c.cleanup()
		return err
	}

	c.logger.Info("debug session ready", "program", dll)
	return nil
}

func (c *Client) buildProject(projectPath string) (string, error) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return "", err
	}

	c.logger.Info("building project", "path", absPath)

	build := exec.Command("dotnet", "build", absPath, "-c", "Debug", "-o", filepath.Join(absPath, "bin", "Debug", "output"))
	build.Stderr = os.Stderr
	build.Stdout = os.Stderr

	if err := build.Run(); err != nil {
		return "", fmt.Errorf("dotnet build: %w", err)
	}

	projectName := filepath.Base(absPath)
	dll := filepath.Join(absPath, "bin", "Debug", "output", projectName+".dll")

	if _, err := os.Stat(dll); err != nil {
		return "", fmt.Errorf("built DLL not found at %s: %w", dll, err)
	}

	c.logger.Info("build complete", "dll", dll)
	return dll, nil
}

func (c *Client) LastProject() string {
	return c.lastProject
}

func (c *Client) Session() *Session {
	return c.session
}

func (c *Client) IsConnected() bool {
	return c.session != nil
}

func (c *Client) Disconnect() error {
	if c.session != nil {
		c.session.Disconnect()
	}
	c.cleanup()
	return nil
}

func (c *Client) cleanup() {
	if c.process != nil && c.process.Process != nil {
		_ = c.process.Process.Kill()
		_ = c.process.Wait()
		c.logger.Info("netcoredbg process terminated")
	}
	c.session = nil
	c.process = nil
}
