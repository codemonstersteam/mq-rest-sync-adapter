package team.codemonsters.refactoringexample.walletBalance


import team.codemonsters.refactoringexample.util.JsonUtil

data class HttpRequestEnvelope(
    val correlationId: String,
    val method: String,
    val service: String,
    val operation: String,
    val headers: Map<String, String>,
    val body: String?
) {
    companion object {
        fun parseJsonString(jsonString: String) =
            Result.runCatching {
                JsonUtil.fromString(jsonString, HttpRequestEnvelope::class.java)
            }
    }

}
