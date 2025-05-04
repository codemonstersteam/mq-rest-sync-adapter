package team.codemonsters.refactoringexample.mq

import org.apache.commons.text.StringSubstitutor
import team.codemonsters.refactoringexample.configuration.RestConfiguration
import team.codemonsters.refactoringexample.configuration.RestServiceCfg
import team.codemonsters.refactoringexample.exceptions.ServiceCommunicationException

data class RestServiceRequest private constructor(
    val restCfg: RestServiceCfg,
    val requestEnvelope: HttpRequestEnvelope,
    val operationUrl: String
) {

    companion object {
        fun emerge(
            httpRequestEnvelope: Result<HttpRequestEnvelope>,
            restConfiguration: RestConfiguration
        ): Result<RestServiceRequest> =
            httpRequestEnvelope.fold(
                onSuccess = { httpRequest ->
                    val restCfg = restConfiguration.configs[httpRequest.service]
                    val substitutor = StringSubstitutor(httpRequest.headers)
                    val validatedRequest = validateRequest(httpRequest, restConfiguration, restCfg)

                    validatedRequest.fold(
                        onSuccess = {
                            Result.success(
                                RestServiceRequest(
                                    restCfg!!,
                                    httpRequest,
                                    substitutor.replace(restCfg.operations[httpRequest.operation]!!)
                                )
                            )
                        },
                        onFailure = { error -> Result.failure(error) }
                    )
                },
                onFailure = { error -> Result.failure(error) }
            )

        private fun validateRequest(
            httpRequestEnvelope: HttpRequestEnvelope,
            restConfiguration: RestConfiguration,
            restCfg: RestServiceCfg?
        ): Result<HttpRequestEnvelope>
        {
            if (!restConfiguration.configs.containsKey(httpRequestEnvelope.service)) {
                return Result.failure(
                    ServiceCommunicationException(
                        httpRequestEnvelope.service,
                        "Не найдена конфигурация для сервиса: ${httpRequestEnvelope.service}"
                    )
                )
            }

            if (!restCfg!!.operations.containsKey(httpRequestEnvelope.operation)) {
                return Result.failure(
                    ServiceCommunicationException(
                        httpRequestEnvelope.service,
                        "Не найдена операция: ${httpRequestEnvelope.operation} для сервиса ${httpRequestEnvelope.service}"
                    )
                )
            }
            return Result.success(httpRequestEnvelope)
        }
    }
}