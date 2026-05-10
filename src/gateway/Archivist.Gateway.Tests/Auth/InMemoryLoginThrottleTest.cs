using Archivist.Gateway.Application.Auth.Services.Defaults;

namespace Archivist.Gateway.Tests.Auth;

/// <summary>
/// Unit tests for <see cref="InMemoryLoginThrottle"/>.
/// </summary>
public sealed class InMemoryLoginThrottleTest
{
    [Fact]
    public void IsThrottled_NoFailures_ReturnsFalse()
    {
        var throttle = new InMemoryLoginThrottle();

        Assert.False(throttle.IsThrottled("1.2.3.4"));
    }

    [Fact]
    public void IsThrottled_BelowPerIpLimit_ReturnsFalse()
    {
        var throttle = new InMemoryLoginThrottle();

        for (var i = 0; i < 9; i++)
        {
            throttle.RecordFailure("1.2.3.4");
        }

        Assert.False(throttle.IsThrottled("1.2.3.4"));
    }

    [Fact]
    public void IsThrottled_AtPerIpLimit_ReturnsTrue()
    {
        var throttle = new InMemoryLoginThrottle();

        for (var i = 0; i < 10; i++)
        {
            throttle.RecordFailure("1.2.3.4");
        }

        Assert.True(throttle.IsThrottled("1.2.3.4"));
    }

    [Fact]
    public void IsThrottled_DifferentIp_NotAffectedByOtherIpLimit()
    {
        var throttle = new InMemoryLoginThrottle();

        for (var i = 0; i < 10; i++)
        {
            throttle.RecordFailure("1.2.3.4");
        }

        // A different IP should not be per-IP throttled.
        Assert.False(throttle.IsThrottled("9.9.9.9"));
    }

    [Fact]
    public void IsThrottled_GlobalLimitReached_ThrottlesAllIps()
    {
        var throttle = new InMemoryLoginThrottle();

        // Spread 50 failures across many IPs to avoid per-IP limit.
        for (var i = 0; i < 50; i++)
        {
            throttle.RecordFailure($"10.0.0.{i}");
        }

        // A fresh IP should be globally throttled.
        Assert.True(throttle.IsThrottled("192.168.1.1"));
    }

    [Fact]
    public void RecordSuccess_ResetsPerIpCounter()
    {
        var throttle = new InMemoryLoginThrottle();

        for (var i = 0; i < 10; i++)
        {
            throttle.RecordFailure("1.2.3.4");
        }

        Assert.True(throttle.IsThrottled("1.2.3.4"));

        throttle.RecordSuccess("1.2.3.4");

        Assert.False(throttle.IsThrottled("1.2.3.4"));
    }
}