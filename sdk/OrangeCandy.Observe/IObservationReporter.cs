namespace OrangeCandy.Observe;

public interface IObservationReporter
{
    void Enqueue(ObservationEvent evt);
}
