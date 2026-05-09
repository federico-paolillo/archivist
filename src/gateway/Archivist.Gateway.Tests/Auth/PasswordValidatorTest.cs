using Archivist.Gateway.Application.Auth.Services.Defaults;

namespace Archivist.Gateway.Tests.Auth;

public sealed class PasswordValidatorTest
{
    private readonly PasswordValidator _validator = new();

    [Fact]
    public void IsValid_WithExactly2048PrintableAsciiCharacters_ReturnsTrue()
    {
        var password = new string('a', PasswordValidator.RequiredLength);

        Assert.True(_validator.IsValid(password));
    }

    [Fact]
    public void IsValid_WithAllPrintableAsciiCharacters_ReturnsTrue()
    {
        // Build a 2048-char string cycling through all printable ASCII (0x20–0x7E).
        var printableRange = Enumerable.Range(0x20, 0x7F - 0x20).Select(i => (char)i).ToArray();
        var chars = new char[PasswordValidator.RequiredLength];
        for (var i = 0; i < chars.Length; i++)
        {
            chars[i] = printableRange[i % printableRange.Length];
        }

        Assert.True(_validator.IsValid(new string(chars)));
    }

    [Fact]
    public void IsValid_WithPasswordShorterThan2048_ReturnsFalse()
    {
        var password = new string('a', PasswordValidator.RequiredLength - 1);

        Assert.False(_validator.IsValid(password));
    }

    [Fact]
    public void IsValid_WithPasswordLongerThan2048_ReturnsFalse()
    {
        var password = new string('a', PasswordValidator.RequiredLength + 1);

        Assert.False(_validator.IsValid(password));
    }

    [Fact]
    public void IsValid_WithEmptyString_ReturnsFalse()
    {
        Assert.False(_validator.IsValid(string.Empty));
    }

    [Fact]
    public void IsValid_WithNonPrintableAsciiCharacter_ReturnsFalse()
    {
        // Insert a tab character (0x09) which is not printable ASCII.
        var chars = new char[PasswordValidator.RequiredLength];
        Array.Fill(chars, 'a');
        chars[0] = '\t';

        Assert.False(_validator.IsValid(new string(chars)));
    }

    [Fact]
    public void IsValid_WithControlCharacter_ReturnsFalse()
    {
        // Newline (0x0A) is below printable ASCII range.
        var chars = new char[PasswordValidator.RequiredLength];
        Array.Fill(chars, 'a');
        chars[500] = '\n';

        Assert.False(_validator.IsValid(new string(chars)));
    }

    [Fact]
    public void IsValid_WithDeleteCharacter_ReturnsFalse()
    {
        // DEL (0x7F) is above printable ASCII range.
        var chars = new char[PasswordValidator.RequiredLength];
        Array.Fill(chars, 'a');
        chars[100] = (char)0x7F;

        Assert.False(_validator.IsValid(new string(chars)));
    }

    [Fact]
    public void IsValid_WithSpaceCharacter_ReturnsTrue()
    {
        // Space (0x20) is the lowest printable ASCII character and must be accepted.
        var chars = new char[PasswordValidator.RequiredLength];
        Array.Fill(chars, ' ');

        Assert.True(_validator.IsValid(new string(chars)));
    }

    [Fact]
    public void IsValid_WithTildeCharacter_ReturnsTrue()
    {
        // Tilde (0x7E) is the highest printable ASCII character and must be accepted.
        var chars = new char[PasswordValidator.RequiredLength];
        Array.Fill(chars, '~');

        Assert.True(_validator.IsValid(new string(chars)));
    }
}