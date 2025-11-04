import "@typespec/http";

using TypeSpec.Http;
@service(#{ title: "DataCosmos Backend Challenge Project" })
@server("{any}", "Any endpoint", {
  any?: string = "https://127.0.0.1:8080",
})
@server("http://127.0.0.1:8080", "localhost endpoint")

namespace DataCosmosChallenge;

model DataPointModel {
  @encode(DurationKnownEncoding.ISO8601)
  time: duration;
  value: float32;
}

@error
model Error {
  @statusCode
  code: 500;
  message: string;
}

@route("/data-point")
@tag("Data Point")
interface DataPoint {
  /** Query Data Point */
  @get query(@query start?: duration, @query until?: duration): DataPointModel[] | Error;
}