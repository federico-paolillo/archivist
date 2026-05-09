namespace Archivist.Gateway.Application.Persistence;

/// <summary>
/// Generates ULID identifiers for persistence records.
/// </summary>
public interface IUlidGenerator
{
    /// <summary>
    /// Creates a new ULID string.
    /// </summary>
    string NewId();
}