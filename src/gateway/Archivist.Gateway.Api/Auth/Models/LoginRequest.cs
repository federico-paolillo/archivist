namespace Archivist.Gateway.Api.Auth.Models;

/// <summary>
/// Request body for <c>POST /login</c>.
/// </summary>
public sealed class LoginRequest
{
    /// <summary>The plaintext login password. Must be exactly 2048 printable ASCII characters.</summary>
    public string? Password { get; set; }
}