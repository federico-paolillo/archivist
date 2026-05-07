namespace Archivist.Gateway.Application.Persistence.Entities;

/// <summary>
/// Represents the fixed personal Archivist user.
/// </summary>
public sealed class UserEntity
{
    /// <summary>
    /// Gets or sets the user ULID.
    /// </summary>
    public required string Id { get; set; }

    /// <summary>
    /// Gets or sets the mapped Telegram sender user id.
    /// </summary>
    public long? TelegramUserId { get; set; }

    /// <summary>
    /// Gets or sets the auth-owned Argon2id PHC password hash.
    /// </summary>
    public string? PasswordHash { get; set; }
}