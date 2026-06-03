using Archivist.Gateway.Application.Persistence.Entities;

using Microsoft.EntityFrameworkCore;

namespace Archivist.Gateway.Application.Persistence;

/// <summary>
/// EF Core SQLite context for Archivist metadata, queue, and notification state.
/// </summary>
public sealed class ArchivistDbContext(DbContextOptions<ArchivistDbContext> options) : DbContext(options)
{
    /// <summary>
    /// Gets persisted users.
    /// </summary>
    public DbSet<UserEntity> Users => Set<UserEntity>();

    /// <summary>
    /// Gets persisted articles.
    /// </summary>
    public DbSet<ArticleEntity> Articles => Set<ArticleEntity>();

    /// <summary>
    /// Gets persisted jobs.
    /// </summary>
    public DbSet<JobEntity> Jobs => Set<JobEntity>();

    /// <summary>
    /// Gets persisted notifications.
    /// </summary>
    public DbSet<NotificationEntity> Notifications => Set<NotificationEntity>();

    /// <inheritdoc />
    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        ArgumentNullException.ThrowIfNull(modelBuilder);

        modelBuilder.Entity<UserEntity>(entity =>
        {
            entity.ToTable("users");
            entity.HasKey(x => x.Id);
            entity.HasIndex(x => x.TelegramUserId).IsUnique();
            entity.Property(x => x.Id).HasColumnName("id").HasMaxLength(26);
            entity.Property(x => x.TelegramUserId).HasColumnName("telegram_user_id");
            entity.Property(x => x.PasswordHash).HasColumnName("password_hash");
        });

        modelBuilder.Entity<ArticleEntity>(entity =>
        {
            entity.ToTable("articles", table =>
                table.HasCheckConstraint("ck_articles_status", "status in ('queued', 'ready', 'failed')"));
            entity.HasKey(x => x.Id);
            entity.HasIndex(x => x.UserId);
            entity.Property(x => x.Id).HasColumnName("id").HasMaxLength(26);
            entity.Property(x => x.UserId).HasColumnName("user_id").HasMaxLength(26);
            entity.Property(x => x.OriginalUrl).HasColumnName("original_url").IsRequired();
            entity.Property(x => x.CanonicalUrl).HasColumnName("canonical_url");
            entity.Property(x => x.Title).HasColumnName("title");
            entity.Property(x => x.Status).HasColumnName("status").HasMaxLength(16);
            entity.Property(x => x.ErrorMessage).HasColumnName("error_message");
            entity.Property(x => x.CreatedAt).HasColumnName("created_at");
            entity.HasOne<UserEntity>().WithMany().HasForeignKey(x => x.UserId).OnDelete(DeleteBehavior.Restrict);
        });

        modelBuilder.Entity<JobEntity>(entity =>
        {
            entity.ToTable("jobs", table =>
                table.HasCheckConstraint("ck_jobs_status", "status in ('queued', 'running', 'succeeded', 'failed')"));
            entity.HasKey(x => x.Id);
            entity.HasIndex(x => x.UserId);
            entity.HasIndex(x => x.ArticleId);
            entity.HasIndex(x => x.TelegramUpdateId).IsUnique();
            entity.Property(x => x.Id).HasColumnName("id").HasMaxLength(26);
            entity.Property(x => x.UserId).HasColumnName("user_id").HasMaxLength(26);
            entity.Property(x => x.ArticleId).HasColumnName("article_id").HasMaxLength(26);
            entity.Property(x => x.Type).HasColumnName("type").HasMaxLength(64);
            entity.Property(x => x.Status).HasColumnName("status").HasMaxLength(16);
            entity.Property(x => x.TelegramUpdateId).HasColumnName("telegram_update_id");
            entity.Property(x => x.TelegramChatId).HasColumnName("telegram_chat_id");
            entity.Property(x => x.TelegramMessageId).HasColumnName("telegram_message_id");
            entity.Property(x => x.TelegramUserId).HasColumnName("telegram_user_id");
            entity.Property(x => x.ErrorMessage).HasColumnName("error_message");
            entity.Property(x => x.CreatedAt).HasColumnName("created_at");
            entity.Property(x => x.StartedAt).HasColumnName("started_at");
            entity.Property(x => x.CompletedAt).HasColumnName("completed_at");
            entity.Property(x => x.ExpiresAt).HasColumnName("expires_at");
            entity.Property(x => x.TraceParent).HasColumnName("traceparent");
            entity.Property(x => x.TraceState).HasColumnName("tracestate");
            entity.HasOne<UserEntity>().WithMany().HasForeignKey(x => x.UserId).OnDelete(DeleteBehavior.Restrict);
            entity.HasOne<ArticleEntity>().WithMany().HasForeignKey(x => x.ArticleId).OnDelete(DeleteBehavior.Cascade);
        });

        modelBuilder.Entity<NotificationEntity>(entity =>
        {
            entity.ToTable("notifications", table =>
                table.HasCheckConstraint("ck_notifications_status", "status in ('pending', 'sent', 'failed')"));
            entity.HasKey(x => x.Id);
            entity.HasIndex(x => x.JobId).IsUnique();
            entity.Property(x => x.Id).HasColumnName("id").HasMaxLength(26);
            entity.Property(x => x.JobId).HasColumnName("job_id").HasMaxLength(26);
            entity.Property(x => x.Status).HasColumnName("status").HasMaxLength(16);
            entity.Property(x => x.ErrorMessage).HasColumnName("error_message");
            entity.Property(x => x.CreatedAt).HasColumnName("created_at");
            entity.Property(x => x.SentAt).HasColumnName("sent_at");
            entity.Property(x => x.ExpiresAt).HasColumnName("expires_at");
            entity.HasOne<JobEntity>().WithMany().HasForeignKey(x => x.JobId).OnDelete(DeleteBehavior.Cascade);
        });
    }
}