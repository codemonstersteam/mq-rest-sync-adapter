package team.codemonsters.refactoringexample.walletBalance


import team.codemonsters.refactoringexample.util.JsonUtil

data class HttpRequestEnvelope constructor(
    /** Нужно сгенерить на стороне вызывающей стороны*/
    val correlationId: String,

    /** POST - другие методы пока не работают TODO - добавить GET */
    val method: String,

    /** для обработки - cifrub-processing, для данных - cifrub-client-data*/
    val service: String,

    /** Название операции, например customer-buy-cr
     * Список операций для каждого сервиса предопределён и задаётся в конфигурации перекладчика */
    val operation: String,

    /** RequestId (пока наличие не проверяется, но видимо будет) */
    val headers: Map<String, String>,

    /** JSON в виде String */
    val body: String?
) {
    companion object {
        fun fromJsonString(jsonString: String) =
            Result.runCatching {
                JsonUtil.fromString(jsonString, HttpRequestEnvelope::class.java)
            }
    }

}
