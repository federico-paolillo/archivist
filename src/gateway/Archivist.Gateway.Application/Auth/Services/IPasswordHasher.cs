namespace Archivist.Gateway.Application.Auth.Services;

/// <summary>
/// Hashes and verifies passwords using Argon2id with PHC string format encoding.
/// </summary>
public interface IPasswordHasher
{
    /// <summary>
    /// Hashes the provided password and returns an Argon2id PHC string.
    /// </summary>
    /// <param name="password">The plaintext password to hash. Must be a valid 2048-character printable ASCII string.</param>
    /// <returns>An Argon2id PHC encoded hash string.</returns>
    string Hash(string password);

    /// <summary>
    /// Verifies a plaintext password against a stored Argon2id PHC hash string using constant-time comparison.
    /// </summary>
    /// <param name="password">The plaintext password to verify.</param>
    /// <param name="phcHash">The stored Argon2id PHC hash string.</param>
    /// <returns><c>true</c> if the password matches the hash; otherwise <c>false</c>.</returns>
    bool Verify(string password, string phcHash);
}