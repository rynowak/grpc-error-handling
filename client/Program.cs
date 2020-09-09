using System;
using System.Linq;
using System.Net.Http;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using Grpc.Core;
using Grpc.Net.Client;

namespace client
{
    class Program
    {
        static async Task Main(string[] args)
        {
            AppContext.SetSwitch("System.Net.Http.SocketsHttpHandler.Http2UnencryptedSupport", true);

            using var channel = GrpcChannel.ForAddress("http://localhost:5000", new GrpcChannelOptions()
            {
                HttpHandler = new Handler(new SocketsHttpHandler()),
                DisposeHttpClient = true,
            });
            var client =  new GrpcGreeter.Greeter.GreeterClient(channel);

            try
            {
                var name = args.Length == 1 ? args[0] : "Pranav";
                var reply = await client.SayHelloAsync(new GrpcGreeter.HelloRequest { Name = name, });
                Console.WriteLine("Greeting: " + reply.Message);
            }
            catch (RpcException ex)
            {
                Console.WriteLine($"Status Code: {ex.Status.StatusCode}");
                Console.WriteLine($"Status Detail: {ex.Status.Detail}");

                Console.WriteLine($"Trailer Count: {ex.Trailers.Count}");
                foreach (var kvp in ex.Trailers)
                {
                    Console.WriteLine($"Trailer: {kvp.Key}: {(kvp.IsBinary ? "(binary data)" : kvp.Value)}");
                }
            }
        }

        class Handler : DelegatingHandler
        {
            public Handler(HttpMessageHandler inner)
                : base(inner)
            {
            }

            protected async override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
            {
                var response = await base.SendAsync(request, cancellationToken);

                Console.WriteLine($"Status Code: {response.StatusCode}");
                foreach (var header in response.Headers)
                {
                    Console.WriteLine($"Header: {header.Key}: {string.Join(", ", header.Value)}");
                }
                if (response.Content != null)
                {
                    foreach (var header in response.Content.Headers)
                    {
                        Console.WriteLine($"Header: {header.Key}: {string.Join(", ", header.Value)}");
                    }
                }
                
                foreach (var header in response.TrailingHeaders)
                {
                    Console.WriteLine($"Trailer: {header.Key}: {string.Join(", ", header.Value)}");
                }

                if (response.Headers.TryGetValues("grpc-status-details-bin", out var values))
                {
                    var value = values.First();
                    var status = Google.Rpc.Status.Parser.ParseFrom(Convert.FromBase64String(value));

                    Console.WriteLine($"Status: {status.Code}");
                    Console.WriteLine($"Status: {status.Message}");
                    Console.WriteLine($"Status: {status.Details.Count}");
                    foreach (var detail in status.Details)
                    {
                        Console.WriteLine($"Detail Type: {detail.TypeUrl}");

                        foreach (var type in Google.Rpc.ErrorDetailsReflection.Descriptor.MessageTypes)
                        {
                            if (detail.Is(type))
                            {
                                var parsed = type.Parser.ParseFrom(detail.Value);
                                Console.WriteLine($"Parsed Detail Type: {parsed.GetType()}");
                            }
                        }
                    }
                }

                return response;
            }
        }
    }
}
