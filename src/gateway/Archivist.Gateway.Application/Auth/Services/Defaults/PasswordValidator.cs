namespace Archivist.Gateway.Application.Auth.Services.Defaults;

/// <summary>
/// Validates that a password is exactly 2048 printable ASCII characters (0x20–0x7E).
/// </summary>
public sealed class PasswordValidator : IPasswordValidator
{
    /// <summary>The required exact length for a valid password.</summary>
    public const int RequiredLength = 2048;

    public bool IsValid(string password)
    {
        ArgumentNullException.ThrowIfNull(password);

        if (password.Length != RequiredLength)
        {
            return false;
        }

        foreach (var ch in password)
        {
            // Printable ASCII: 0x20 (space) through 0x7E (tilde) inclusive.
            if (ch < 0x20 || ch > 0x7E)
            {
                return false;
            }
        }

        return true;
    }
}