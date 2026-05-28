namespace Archivist.Gateway.Application.ArticleArtifacts;

/// <summary>
/// Represents an operational failure while reading an article artifact.
/// </summary>
public sealed class ArticleArtifactReadException : Exception
{
    /// <summary>
    /// Initializes a new instance of the <see cref="ArticleArtifactReadException"/> class.
    /// </summary>
    public ArticleArtifactReadException()
    {
    }

    /// <summary>
    /// Initializes a new instance of the <see cref="ArticleArtifactReadException"/> class.
    /// </summary>
    public ArticleArtifactReadException(string message)
        : base(message)
    {
    }

    /// <summary>
    /// Initializes a new instance of the <see cref="ArticleArtifactReadException"/> class.
    /// </summary>
    public ArticleArtifactReadException(string message, Exception innerException)
        : base(message, innerException)
    {
    }
}