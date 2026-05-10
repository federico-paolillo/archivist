namespace Archivist.Gateway.Application.Auth.Services;

/// <summary>
/// Validates that a password meets the required format constraints.
/// </summary>
public interface IPasswordValidator
{
    /// <summary>
    /// Validates that the provided password is exactly 2048 printable ASCII characters (0x20–0x7E).
    /// </summary>
    /// <param name="password">The password to validate.</param>
    /// <returns><c>true</c> if the password is valid; otherwise <c>false</c>.</returns>
    bool IsValid(string password);
}