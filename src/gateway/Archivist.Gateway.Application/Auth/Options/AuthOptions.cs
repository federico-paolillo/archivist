namespace Archivist.Gateway.Application.Auth.Options;

/// <summary>
/// Configuration options for authentication bootstrap and password management.
/// </summary>
public sealed class AuthOptions
{
    public const string SectionName = "Auth";

    /// <summary>
    /// The one-time bootstrap password used to initialize <c>password_hash</c> for the personal user.
    /// Required only when the personal user row has no stored password hash.
    /// Must be exactly 2048 printable ASCII characters.
    /// Must never be logged or persisted in plaintext.
    /// </summary>
    public string? BootstrapPassword { get; set; }

    /// <summary>
    /// The path to the SQLite database file.
    /// </summary>
    public string? SqlitePath { get; set; }
}