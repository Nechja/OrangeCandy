package dap

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	projectFile, projectDir, err := resolveProjectFile(absPath)
	if err != nil {
		return "", err
	}

	outputDir := filepath.Join(projectDir, "bin", "Debug", "output")
	c.logger.Info("building project", "file", projectFile, "output", outputDir)

	build := exec.Command("dotnet", "build", projectFile, "-c", "Debug", "-o", outputDir)
	build.Stderr = os.Stderr
	build.Stdout = os.Stderr

	if err := build.Run(); err != nil {
		return "", fmt.Errorf("dotnet build: %w", err)
	}

	projectName := strings.TrimSuffix(filepath.Base(projectFile), filepath.Ext(projectFile))
	dll := filepath.Join(outputDir, projectName+".dll")

	if _, err := os.Stat(dll); err != nil {
		return "", fmt.Errorf("built DLL not found at %s: %w", dll, err)
	}

	c.logger.Info("build complete", "dll", dll)
	return dll, nil
}

// resolveProjectFile returns the build target (a file path) and its parent
// directory. The input may be either a project/solution file or a directory.
// When given a directory, it picks the single project/solution file inside;
// if multiple exist, it returns an error listing the candidates.
func resolveProjectFile(absPath string) (string, string, error) {
	info, err := os.Stat(absPath)
	if err != nil {
		return "", "", fmt.Errorf("stat %s: %w", absPath, err)
	}

	if !info.IsDir() {
		if !isProjectExt(filepath.Ext(absPath)) {
			return "", "", fmt.Errorf("unsupported file extension %q (expected .csproj, .fsproj, .vbproj, .sln, or .slnx)", filepath.Ext(absPath))
		}
		return absPath, filepath.Dir(absPath), nil
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return "", "", fmt.Errorf("read dir %s: %w", absPath, err)
	}

	var candidates []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if isProjectExt(filepath.Ext(e.Name())) {
			candidates = append(candidates, filepath.Join(absPath, e.Name()))
		}
	}

	switch len(candidates) {
	case 0:
		return "", "", fmt.Errorf("no project or solution file found in %s", absPath)
	case 1:
		return candidates[0], absPath, nil
	default:
		return "", "", fmt.Errorf("multiple project/solution files in %s: %v — pass a specific file path instead", absPath, candidates)
	}
}

func isProjectExt(ext string) bool {
	switch ext {
	case ".csproj", ".fsproj", ".vbproj", ".sln", ".slnx":
		return true
	}
	return false
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
