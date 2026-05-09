using System.Security.Cryptography;
using System.Text;

using Konscious.Security.Cryptography;

namespace Archivist.Gateway.Application.Auth.Services.Defaults;

/// <summary>
/// Hashes and verifies passwords using Argon2id with PHC string format encoding.
/// Parameters: m=19456, t=2, p=1 as required by AUTHN SPEC REQ-005.
/// </summary>
public sealed class Argon2idPasswordHasher : IPasswordHasher
{
    private const int MemorySize = 19456; // KiB
    private const int Iterations = 2;
    private const int DegreeOfParallelism = 1;
    private const int SaltLength = 16; // bytes
    private const int HashLength = 32; // bytes

    public string Hash(string password)
    {
        ArgumentNullException.ThrowIfNull(password);

        var passwordBytes = Encoding.UTF8.GetBytes(password);
        var salt = RandomNumberGenerator.GetBytes(SaltLength);

        var hash = ComputeHash(passwordBytes, salt);

        return EncodePHC(salt, hash);
    }

    public bool Verify(string password, string phcHash)
    {
        ArgumentNullException.ThrowIfNull(password);
        ArgumentNullException.ThrowIfNull(phcHash);

        if (!TryDecodePHC(phcHash, out var salt, out var expectedHash))
        {
            return false;
        }

        var passwordBytes = Encoding.UTF8.GetBytes(password);
        var actualHash = ComputeHash(passwordBytes, salt!);

        // Constant-time comparison to prevent timing attacks.
        return CryptographicOperations.FixedTimeEquals(actualHash, expectedHash!);
    }

    private static byte[] ComputeHash(byte[] passwordBytes, byte[] salt)
    {
        using var argon2 = new Argon2id(passwordBytes);
        argon2.Salt = salt;
        argon2.MemorySize = MemorySize;
        argon2.Iterations = Iterations;
        argon2.DegreeOfParallelism = DegreeOfParallelism;
        return argon2.GetBytes(HashLength);
    }

    private static string EncodePHC(byte[] salt, byte[] hash)
    {
        // PHC format: $argon2id$v=19$m=<mem>,t=<iter>,p=<para>$<salt_base64>$<hash_base64>
        var saltBase64 = Convert.ToBase64String(salt).TrimEnd('=');
        var hashBase64 = Convert.ToBase64String(hash).TrimEnd('=');
        return $"$argon2id$v=19$m={MemorySize},t={Iterations},p={DegreeOfParallelism}${saltBase64}${hashBase64}";
    }

    private static bool TryDecodePHC(
        string phcHash,
        out byte[]? salt,
        out byte[]? hash)
    {
        salt = null;
        hash = null;

        // Expected: $argon2id$v=19$m=<mem>,t=<iter>,p=<para>$<salt_base64>$<hash_base64>
        var parts = phcHash.Split('$');

        // parts[0] = "" (empty before first $)
        // parts[1] = "argon2id"
        // parts[2] = "v=19"
        // parts[3] = "m=...,t=...,p=..."
        // parts[4] = salt base64
        // parts[5] = hash base64
        if (parts.Length != 6)
        {
            return false;
        }

        if (!string.IsNullOrEmpty(parts[0]) || parts[1] != "argon2id")
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

    private static string PadBase64(string base64)
    {
        // Restore stripped padding.
        var mod = base64.Length % 4;
        return mod switch
        {
            0 => base64,
            2 => base64 + "==",
            3 => base64 + "=",
            _ => base64, // mod == 1 is invalid base64 but let Convert.FromBase64String handle it
        };
    }
}