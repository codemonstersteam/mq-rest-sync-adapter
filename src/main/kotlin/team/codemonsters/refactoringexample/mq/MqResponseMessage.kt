package team.codemonsters.refactoringexample.mq

import team.codemonsters.refactoringexample.util.JsonUtil

data class MqResponseMessage(
    val correlationId: String,
    val operation: String,
    val payload: String
) {

    companion object {
        fun emerge(data: Result<HttpResponseEnvelope>): Result<MqResponseMessage> =
            data.fold(
                onSuccess = { restResponse ->
                    Result.success(
                        MqResponseMessage(
                            restResponse.correlationId,
                            restResponse.operation ?: "",
                            JsonUtil.toJson(restResponse)
                        )
                    )
                },
                onFailure = { error -> Result.failure(error) }
            )


        fun emerge(
            correlationId: String,
            url: String,
            request: Result<HttpResponseEnvelope>
        ): Result<MqResponseMessage> =
            request.fold(
                onSuccess = { success ->
                    Result.success(
                        MqResponseMessage(
                            correlationId,
                            url,
                            success.body ?: ""
                        )
                    )
                },
                onFailure = { error -> Result.failure(error) }
            )


        fun emerge(correlationId: String, url: String, body: String): MqResponseMessage {
            return MqResponseMessage(
                correlationId,
                url,
                body
            )
        }

        fun emerge(correlationId: String, url: String, ex: Throwable): MqResponseMessage {
            return MqResponseMessage(
                correlationId,
                url,
                ex.toString()
            )
        }
    }
}
