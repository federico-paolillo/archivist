using Archivist.Gateway.Application.Auth.Services.Defaults;

namespace Archivist.Gateway.Tests.Auth;

public sealed class Argon2idPasswordHasherTest
{
    private readonly Argon2idPasswordHasher _hasher = new();

    private static string ValidPassword() => new('a', PasswordValidator.RequiredLength);

    [Fact]
    public void Hash_ReturnsArgon2idPHCString()
    {
        var password = ValidPassword();

        var hash = _hasher.Hash(password);

        Assert.StartsWith("$argon2id$", hash);
    }

    [Fact]
    public void Hash_ProducesDifferentHashesForSamePassword()
    {
        // Each call uses a new random salt so hashes must differ.
        var password = ValidPassword();

        var hash1 = _hasher.Hash(password);
        var hash2 = _hasher.Hash(password);

        Assert.NotEqual(hash1, hash2);
    }

    [Fact]
    public void Verify_WithCorrectPassword_ReturnsTrue()
    {
        var password = ValidPassword();
        var hash = _hasher.Hash(password);

        Assert.True(_hasher.Verify(password, hash));
    }

    [Fact]
    public void Verify_WithWrongPassword_ReturnsFalse()
    {
        var password = ValidPassword();
        var hash = _hasher.Hash(password);
        var wrongPassword = new string('b', PasswordValidator.RequiredLength);

        Assert.False(_hasher.Verify(wrongPassword, hash));
    }

    [Fact]
    public void Verify_WithCorruptedHash_ReturnsFalse()
    {
        var password = ValidPassword();

        Assert.False(_hasher.Verify(password, "not-a-phc-string"));
    }

    [Fact]
    public void Verify_WithTruncatedPHCString_ReturnsFalse()
    {
        var password = ValidPassword();
        var hash = _hasher.Hash(password);
        var truncated = hash[..10];

        Assert.False(_hasher.Verify(password, truncated));
    }

    [Fact]
    public void Hash_PHCStringContainsExpectedParameters()
    {
        var password = ValidPassword();
        var hash = _hasher.Hash(password);

        // PHC format: $argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash>
        Assert.Contains("v=19", hash);
        Assert.Contains("m=19456", hash);
        Assert.Contains("t=2", hash);
        Assert.Contains("p=1", hash);
    }
}