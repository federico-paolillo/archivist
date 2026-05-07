using System.Security.Cryptography;

namespace Archivist.Gateway.Application.Persistence.Defaults;

/// <summary>
/// Generates Crockford Base32 ULID strings using system randomness.
/// </summary>
public sealed class RandomUlidGenerator(TimeProvider timeProvider) : IUlidGenerator
{
    private const string Alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ";

    /// <inheritdoc />
    public string NewId()
    {
        Span<byte> bytes = stackalloc byte[16];
        var timestamp = timeProvider.GetUtcNow().ToUnixTimeMilliseconds();

        bytes[0] = (byte)(timestamp >> 40);
        bytes[1] = (byte)(timestamp >> 32);
        bytes[2] = (byte)(timestamp >> 24);
        bytes[3] = (byte)(timestamp >> 16);
        bytes[4] = (byte)(timestamp >> 8);
        bytes[5] = (byte)timestamp;
        RandomNumberGenerator.Fill(bytes[6..]);

        Span<char> chars = stackalloc char[26];
        chars[0] = Alphabet[(bytes[0] & 224) >> 5];
        chars[1] = Alphabet[bytes[0] & 31];
        chars[2] = Alphabet[(bytes[1] & 248) >> 3];
        chars[3] = Alphabet[((bytes[1] & 7) << 2) | ((bytes[2] & 192) >> 6)];
        chars[4] = Alphabet[(bytes[2] & 62) >> 1];
        chars[5] = Alphabet[((bytes[2] & 1) << 4) | ((bytes[3] & 240) >> 4)];
        chars[6] = Alphabet[((bytes[3] & 15) << 1) | ((bytes[4] & 128) >> 7)];
        chars[7] = Alphabet[(bytes[4] & 124) >> 2];
        chars[8] = Alphabet[((bytes[4] & 3) << 3) | ((bytes[5] & 224) >> 5)];
        chars[9] = Alphabet[bytes[5] & 31];
        chars[10] = Alphabet[(bytes[6] & 248) >> 3];
        chars[11] = Alphabet[((bytes[6] & 7) << 2) | ((bytes[7] & 192) >> 6)];
        chars[12] = Alphabet[(bytes[7] & 62) >> 1];
        chars[13] = Alphabet[((bytes[7] & 1) << 4) | ((bytes[8] & 240) >> 4)];
        chars[14] = Alphabet[((bytes[8] & 15) << 1) | ((bytes[9] & 128) >> 7)];
        chars[15] = Alphabet[(bytes[9] & 124) >> 2];
        chars[16] = Alphabet[((bytes[9] & 3) << 3) | ((bytes[10] & 224) >> 5)];
        chars[17] = Alphabet[bytes[10] & 31];
        chars[18] = Alphabet[(bytes[11] & 248) >> 3];
        chars[19] = Alphabet[((bytes[11] & 7) << 2) | ((bytes[12] & 192) >> 6)];
        chars[20] = Alphabet[(bytes[12] & 62) >> 1];
        chars[21] = Alphabet[((bytes[12] & 1) << 4) | ((bytes[13] & 240) >> 4)];
        chars[22] = Alphabet[((bytes[13] & 15) << 1) | ((bytes[14] & 128) >> 7)];
        chars[23] = Alphabet[(bytes[14] & 124) >> 2];
        chars[24] = Alphabet[((bytes[14] & 3) << 3) | ((bytes[15] & 224) >> 5)];
        chars[25] = Alphabet[bytes[15] & 31];

        return new string(chars);
    }
}