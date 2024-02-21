package team.codemonsters.refactoringexample.walletBalance

import org.apache.commons.text.StringSubstitutor
import team.codemonsters.refactoringexample.configuration.RestServiceCfg

data class RestServiceRequest private constructor(
    val restCfg: RestServiceCfg,
    val requestEnvelope: HttpRequestEnvelope,
    val operationUrl: String
) {

    companion object {
        fun of(
            validRequest: ValidRequest
        ): RestServiceRequest {
            val httpRequest = validRequest.unvalidatedRequest.httpRequestEnvelope
            val restCfg = validRequest.unvalidatedRequest.restConfiguration.configs[httpRequest.service]!!
            val substitutor = StringSubstitutor(httpRequest.headers)
            return RestServiceRequest(
                restCfg,
                httpRequest,
                substitutor.replace(restCfg.operations[httpRequest.operation]!!)
            )
        }

    }
}