package team.codemonsters.refactoringexample.walletBalance

import org.apache.commons.text.StringSubstitutor
import java.net.URI

data class ValidRequest private constructor(
    val unvalidatedRequest: UnvalidatedRequest,
    val balanceRequest: PsGetBalanceRequest,
    val httpRequestEnvelope: HttpRequestEnvelope,
    val serviceUri: URI
) {

    fun operationUrl(): String =
        StringSubstitutor(httpRequestEnvelope.headers)
            .replace(serviceOperation())

    private fun serviceOperation() = configuration().operations[httpRequestEnvelope.operation]!!


    fun configuration() = unvalidatedRequest.restConfiguration.configs[httpRequestEnvelope.service]!!

    companion object {

        fun emerge(
            unvalidatedRequest: UnvalidatedRequest,
        ): Result<ValidRequest> =
            HttpRequestEnvelopeDto.parseJsonString(
                unvalidatedRequest.payload.toString()
            )
                .flatMap {
                    HttpRequestEnvelope.emerge(
                        it,
                        unvalidatedRequest.restConfiguration
                    )
                }
                .flatMap { validateRestRequest(unvalidatedRequest, it) }

        private fun validateRestRequest(
            unvalidatedRequest: UnvalidatedRequest,
            httpRequestEnvelope: HttpRequestEnvelope,
        ) : Result<ValidRequest> {
            val balanceRequest = PsGetBalanceRequest.emerge(
                httpRequestEnvelope.body,
                unvalidatedRequest.requestLifetime
            )
            val uri =  createUri(unvalidatedRequest, httpRequestEnvelope)
            return Result.zip(balanceRequest, uri)
                .map{
                    ValidRequest(unvalidatedRequest, it.first, httpRequestEnvelope, it.second)
                }
        }

        private fun createUri(
            unvalidatedRequest: UnvalidatedRequest,
            httpRequestEnvelope: HttpRequestEnvelope,
        ): Result<URI> {
            val restCfg = unvalidatedRequest.restConfiguration.configs[httpRequestEnvelope.service]!!
            val operationUrl =
                StringSubstitutor(httpRequestEnvelope.headers).replace(restCfg.operations[httpRequestEnvelope.operation]!!)
            val uri = restCfg.url + operationUrl
            return Result.runCatching {
                URI.create(uri) }
                .recoverCatching {
                    throw RuntimeException("Failed to create RestServiceRequest cause the url {$uri} is not valid", it)
                }
        }

    }

}
