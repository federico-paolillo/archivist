using System.Security.Cryptography;
using System.Text;

using Konscious.Security.Cryptography;

namespace Archivist.Gateway.Application.Auth.Services.Defaults;

/// <summary>
/// Hashes and verifies passwords using Argon2id with PHC string format encoding.
/// Parameters: m=19456, t=2, p=1 (minimum from SPEC REQ-005).
/// </summary>
public sealed class Argon2idPasswordHasher : IPasswordHasher
{
    private const int MemorySize = 19456; // kibibytes
    private const int Iterations = 2;
    private const int DegreeOfParallelism = 1;
    private const int SaltSize = 16; // bytes
    private const int HashSize = 32; // bytes

    public string Hash(string password)
    {
        ArgumentNullException.ThrowIfNull(password);

        var salt = RandomNumberGenerator.GetBytes(SaltSize);
        var hash = ComputeHash(password, salt);

        // PHC format: $argon2id$v=19$m=19456,t=2,p=1$<salt-base64>$<hash-base64>
        var saltB64 = Convert.ToBase64String(salt).TrimEnd('=');
        var hashB64 = Convert.ToBase64String(hash).TrimEnd('=');

        return $"$argon2id$v=19$m={MemorySize},t={Iterations},p={DegreeOfParallelism}${saltB64}${hashB64}";
    }

    public bool Verify(string password, string phcHash)
    {
        ArgumentNullException.ThrowIfNull(password);
        ArgumentNullException.ThrowIfNull(phcHash);

        if (!TryParsePHC(phcHash, out var salt, out var expectedHash))
        {
            return false;
        }

        var actualHash = ComputeHash(password, salt);

        return CryptographicOperations.FixedTimeEquals(actualHash, expectedHash);
    }

    private static byte[] ComputeHash(string password, byte[] salt)
    {
        var passwordBytes = Encoding.UTF8.GetBytes(password);

        using var argon2 = new Argon2id(passwordBytes);
        argon2.Salt = salt;
        argon2.MemorySize = MemorySize;
        argon2.Iterations = Iterations;
        argon2.DegreeOfParallelism = DegreeOfParallelism;

        return argon2.GetBytes(HashSize);
    }

    private static bool TryParsePHC(string phcHash, out byte[] salt, out byte[] hash)
    {
        salt = [];
        hash = [];

        // Expected: $argon2id$v=19$m=<m>,t=<t>,p=<p>$<salt>$<hash>
        var parts = phcHash.Split('$');
        if (parts.Length < 6 || parts[1] != "argon2id")
        {
            return false;
        }

        try
        {
            salt = Convert.FromBase64String(PadBase64(parts[4]));
            hash = Convert.FromBase64String(PadBase64(parts[5]));
            return true;
        }
        catch (FormatException)
        {
            return false;
        }
    }

    private static string PadBase64(string s)
    {
        var padding = (4 - (s.Length % 4)) % 4;
        return s + new string('=', padding);
    }
}