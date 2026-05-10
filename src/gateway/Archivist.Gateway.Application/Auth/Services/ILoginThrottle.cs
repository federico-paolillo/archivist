namespace Archivist.Gateway.Application.Auth.Services;

/// <summary>
/// Tracks failed login attempts and enforces per-IP and global rate limits.
/// </summary>
public interface ILoginThrottle
{
    /// <summary>
    /// Returns <c>true</c> if the given source IP (or the global limit) has exceeded the allowed
    /// number of failed login attempts and the request should be rejected before password verification.
    /// </summary>
    bool IsThrottled(string sourceIp);

    /// <summary>
    /// Records a failed login attempt for the given source IP.
    /// </summary>
    void RecordFailure(string sourceIp);

    /// <summary>
    /// Resets the per-IP and global counters for the given source IP after a successful login.
    /// </summary>
    void RecordSuccess(string sourceIp);
}