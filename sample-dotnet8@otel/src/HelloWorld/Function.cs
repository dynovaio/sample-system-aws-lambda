using System.Threading.Tasks;
using System.Net.Http;
using System.IO;
using System.Collections.Generic;
using System;
using Newtonsoft.Json;
using System.Diagnostics;

using Amazon.Lambda.Core;
using Amazon.Lambda.APIGatewayEvents;

using OpenTelemetry.Trace;
using OpenTelemetry.Logs;
using OpenTelemetry.Metrics;
using OpenTelemetry.Instrumentation.AWSLambda;
using OpenTelemetry;


[assembly: LambdaSerializer(typeof(Amazon.Lambda.Serialization.Json.JsonSerializer))]

namespace HelloWorld
{

    public class Function
    {
        private static readonly ActivitySource dynovaLambdaInstrumentation = new ActivitySource("Dynova.Lambda.Instrumentation", "1.0.0");
        private static readonly HttpClient client = new HttpClient();
        public static TracerProvider tracerProvider;

        static Function()
        {
            AppContext.SetSwitch("System.Net.Http.SocketsHttpHandler.Http2UnencryptedSupport", true);

            tracerProvider = Sdk.CreateTracerProviderBuilder()
                .AddSource("Dynova.Lambda.Instrumentation")
                .AddAWSLambdaConfigurations(options => options.DisableAwsXRayContextExtraction = true)
                .AddOtlpExporter()
                .Build();
        }

        static void CheckCollectorConfig(ILambdaContext context)
        {
            using (var checkCollectorConfigActivity = dynovaLambdaInstrumentation.StartActivity("checkCollectorConfigActivity"))
            {
                // List contents of the /opt directory
                string[] files = Directory.GetFiles("/opt");
                context.Logger.LogLine("Contents of /opt directory:");
                foreach (string file in files)
                {
                    context.Logger.LogLine(file);
                }

                // Read and log the collector configuration file
                string configFilePath = Environment.GetEnvironmentVariable("OPENTELEMETRY_COLLECTOR_CONFIG_FILE");
                if (!string.IsNullOrEmpty(configFilePath))
                {
                    if (File.Exists(configFilePath))
                    {
                        string fileContents = File.ReadAllText(configFilePath);
                        context.Logger.LogLine($"Collector config file found at '{configFilePath}'. Contents:");
                        context.Logger.LogLine($"{fileContents}");
                    }
                    else
                    {
                        context.Logger.LogLine($"Collector config file not found at '{configFilePath}'.");
                    }
                }
                else
                {
                    context.Logger.LogLine("OPENTELEMETRY_COLLECTOR_CONFIG_FILE environment variable is not set.");
                }
            }
        }

        private static async Task<string> GetCallingIP()
        {
            using (var getCallingIPActivity = dynovaLambdaInstrumentation.StartActivity("getCallingIPActivity"))
            {
                client.DefaultRequestHeaders.Accept.Clear();
                client.DefaultRequestHeaders.Add("User-Agent", "AWS Lambda .Net Client");

                var msg = await client.GetStringAsync("http://checkip.amazonaws.com/").ConfigureAwait(continueOnCapturedContext:false);

                return msg.Replace("\n","");
            }
        }

        // Make a function wich calls th poke api and returns the data as a json object from the provided pokemon name
        private static async Task<string> GetPokemonData(string name)
        {
            using (var getPokemonDataActivity = dynovaLambdaInstrumentation.StartActivity("getPokemonDataActivity"))
            {
                getPokemonDataActivity?.SetTag("peer.hostname", "pokeapi.co");
                getPokemonDataActivity?.SetTag("server.address", "pokeapi.co");
                getPokemonDataActivity?.SetTag("server.port", 443);
                getPokemonDataActivity?.SetTag("http.method", "GET");
                getPokemonDataActivity?.SetTag("http.url", $"https://pokeapi.co/api/v2/pokemon/{name}");

                client.DefaultRequestHeaders.Accept.Clear();
                client.DefaultRequestHeaders.Add("User-Agent", "AWS Lambda .Net Client");

                client.DefaultRequestHeaders.Add("Accept", "application/json");

                var response = await client.GetAsync($"https://pokeapi.co/api/v2/pokemon/{name}").ConfigureAwait(continueOnCapturedContext:false);

                getPokemonDataActivity?.SetTag("http.statusCode", (int)response.StatusCode);
                getPokemonDataActivity?.SetTag("httpResponseCode", (int)response.StatusCode);

                if (response.IsSuccessStatusCode)
                {
                    var responseData = await response.Content.ReadAsStringAsync().ConfigureAwait(continueOnCapturedContext:false);
                    var pokemonData = JsonConvert.DeserializeObject(responseData);

                    return pokemonData.ToString();
                }
                else
                {
                    getPokemonDataActivity?.SetStatus(ActivityStatusCode.Error, $"Error: {response.StatusCode} - {response.ReasonPhrase}");
                    getPokemonDataActivity?.SetTag("error", true);
                    getPokemonDataActivity?.SetTag("error.message", $"{response.ReasonPhrase}");
                    getPokemonDataActivity?.SetTag("error.type", "HttpRequestException");

                    throw new Exception($"Error: {response.StatusCode} - {response.ReasonPhrase}");
                }
            }
        }

        public async Task<APIGatewayProxyResponse> FunctionHandler(APIGatewayProxyRequest apigProxyEvent, ILambdaContext context)
        {
            CheckCollectorConfig(context);

            var location = await GetCallingIP();

            // Get name param from query string
            var name = apigProxyEvent.QueryStringParameters != null && apigProxyEvent.QueryStringParameters.ContainsKey("name")
                ? apigProxyEvent.QueryStringParameters["name"]
                : "World";

            var pokemonData = await GetPokemonData(name);

            var body = new Dictionary<string, string>
            {
                { "message", $"Hello {name}!" },
                { "pokemonData", pokemonData },
                { "location", location }
            };

            return new APIGatewayProxyResponse
            {
                Body = JsonConvert.SerializeObject(body),
                StatusCode = 200,
                Headers = new Dictionary<string, string> { { "Content-Type", "application/json" } }
            };
        }
        public async Task<APIGatewayProxyResponse> WrappedFunctionHandler(APIGatewayProxyRequest request, ILambdaContext context)
        {
            return await AWSLambdaWrapper.TraceAsync(tracerProvider, FunctionHandler, request, context);
        }
    }
}
