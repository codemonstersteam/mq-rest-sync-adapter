package team.codemonsters.refactoringexample.walletBalance


import team.codemonsters.refactoringexample.configuration.RestConfiguration

data class HttpRequestEnvelope private constructor(
    val correlationId: String,
    val method: String,
    val service: String,
    val operation: String,
    val headers: Map<String, String>,
    val body: String
) {
    companion object {

        fun emerge(
            envelopeDto: HttpRequestEnvelopeDto,
            restConfiguration: RestConfiguration,
        ): Result<HttpRequestEnvelope> =
            Result.success(envelopeDto)
                .ensure({ bodyIsNotNullAndEmpty(it) }, "Отсутствует тела запроса в конверте")
                .flatMap {
                    validateConfigurationMatchEnvelope(it, restConfiguration)
                }.map {
                    fromDto(it)
                }

        private fun bodyIsNotNullAndEmpty(it: HttpRequestEnvelopeDto) = !it.body.isNullOrEmpty()

        private fun validateConfigurationMatchEnvelope(
            httpRequestEnvelope: HttpRequestEnvelopeDto,
            restConfiguration: RestConfiguration
        ): Result<HttpRequestEnvelopeDto> =
            Result.success(httpRequestEnvelope)
                .ensure(
                    { restConfiguration.configs.containsKey(it.service) },
                    "Не найдена конфигурация для сервиса: ${httpRequestEnvelope.service}"
                )
                .ensure(
                    {
                        serviceOperations(restConfiguration, httpRequestEnvelope).containsKey(
                            httpRequestEnvelope.operation,
                        )
                    },
                    "Не найдена операция: ${httpRequestEnvelope.operation} для сервиса ${httpRequestEnvelope.service}"
                )

        private fun serviceOperations(
            restConfiguration: RestConfiguration,
            httpRequestEnvelope: HttpRequestEnvelopeDto,
        ) = restConfiguration.configs[httpRequestEnvelope.service]!!.operations

        private fun fromDto(envelopeDto: HttpRequestEnvelopeDto) =
            HttpRequestEnvelope(
                correlationId = envelopeDto.correlationId,
                method = envelopeDto.method,
                service = envelopeDto.service,
                operation = envelopeDto.operation,
                headers = envelopeDto.headers,
                body = envelopeDto.body!!
            )

    }

}
