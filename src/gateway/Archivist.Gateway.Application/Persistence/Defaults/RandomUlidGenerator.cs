using NUlid;

namespace Archivist.Gateway.Application.Persistence.Defaults;

/// <summary>
/// Generates Crockford Base32 ULID strings.
/// </summary>
public sealed class RandomUlidGenerator(TimeProvider timeProvider) : IUlidGenerator
{
    /// <inheritdoc />
    public string NewId() => Ulid.NewUlid(timeProvider.GetUtcNow()).ToString();
}
