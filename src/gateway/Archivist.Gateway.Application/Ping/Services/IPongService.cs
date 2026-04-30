using Archivist.Gateway.Application.Ping.Models;

namespace Archivist.Gateway.Application.Ping.Services;

public interface IPingService
{
    Pong Ping();
}