package team.codemonsters.refactoringexample.walletBalance

data class ValidRequest private constructor(
    val unvalidatedRequest: UnvalidatedRequest,
    val balanceRequest: PsGetBalanceRequest,
    val httpRequestEnvelope: HttpRequestEnvelope
) {
    companion object {

        fun emerge(
            unvalidatedRequest: UnvalidatedRequest,
        ): Result<ValidRequest> =
            HttpRequestEnvelope.parseJsonString(unvalidatedRequest.payload.toString())
                .ensure({it.body!=null}, "Отсутствует тела запроса в конверте")
                .flatMap {
                    validateRequestConfiguration(it, unvalidatedRequest)
                }

        private fun validateRequestConfiguration(
            httpRequestEnvelope: HttpRequestEnvelope,
            unvalidatedRequest: UnvalidatedRequest
        ) =
            PsGetBalanceRequest.emerge(httpRequestEnvelope.body!!, unvalidatedRequest.requestLifetime)
                .ensure(
                    { unvalidatedRequest.restConfiguration.configs.containsKey(httpRequestEnvelope.service) },
                    "Не найдена конфигурация для сервиса: ${httpRequestEnvelope.service}"
                )
                .ensure(
                    {
                        unvalidatedRequest.restConfiguration.configs[httpRequestEnvelope.service]!!.operations.containsKey(
                            httpRequestEnvelope.operation,
                        )
                    },
                    "Не найдена операция: ${httpRequestEnvelope.operation} для сервиса ${httpRequestEnvelope.service}"
                )
                .map { ValidRequest(unvalidatedRequest, it, httpRequestEnvelope) }
    }

}
