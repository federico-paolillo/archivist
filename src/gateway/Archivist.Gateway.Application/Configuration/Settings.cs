namespace Archivist.Gateway.Application.Configuration;

/// <summary>
/// Collects Gateway configuration source names, sections, and standalone flat logical keys.
/// </summary>
public static class Settings
{
    /// <summary>
    /// Environment variable prefix accepted by the Gateway process.
    /// </summary>
    public const string EnvironmentVariablePrefix = "ARCHIVIST_";

    /// <summary>
    /// Optional defaults file loaded by the Gateway process.
    /// </summary>
    public const string AppSettingsFile = "appsettings.json";

    public const string TelegramSection = "Telegram";

    public const string DataDirectoryKey = "DATA_DIR";
    public const string SqlitePathKey = "SQLITE_PATH";
    public const string AuthBootstrapPasswordKey = "AUTH_BOOTSTRAP_PASSWORD";
    public const string GatewayPublicHostsKey = "GATEWAY_PUBLIC_HOSTS";
}