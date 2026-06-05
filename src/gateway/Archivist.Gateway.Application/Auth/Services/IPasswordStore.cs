namespace Archivist.Gateway.Application.Auth.Services;

/// <summary>
/// Reads password-bearing user credentials from the persistence layer.
/// </summary>
public interface IPasswordStore
{
    /// <summary>
    /// Returns all stored password-bearing user rows.
    /// </summary>
    Task<IReadOnlyList<PasswordCredential>> GetPasswordCredentialsAsync(CancellationToken ct = default);
}

/// <summary>
/// Represents the user id and Argon2id PHC hash used for password login.
/// </summary>
public sealed record PasswordCredential(string UserId, string PasswordHash);