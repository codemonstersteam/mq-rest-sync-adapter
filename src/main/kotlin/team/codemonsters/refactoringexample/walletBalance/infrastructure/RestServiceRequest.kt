package team.codemonsters.refactoringexample.walletBalance.infrastructure

import org.apache.commons.text.StringSubstitutor
import org.springframework.http.HttpHeaders
import org.springframework.http.MediaType
import team.codemonsters.refactoringexample.configuration.RestServiceCfg
import team.codemonsters.refactoringexample.walletBalance.HttpRequestEnvelope
import team.codemonsters.refactoringexample.walletBalance.ValidRequest
import java.net.URI

data class RestServiceRequest private constructor(
    private val restCfg: RestServiceCfg,
    val requestEnvelope: HttpRequestEnvelope,
    val operationUrl: String,
    val uri: URI,
) {

    fun headers(): HttpHeaders {
        val headers = HttpHeaders()
        headers.set(HttpHeaders.CONTENT_TYPE, CONTENT_TYPE)
        requestEnvelope.headers.forEach { (k, v) ->
            headers.set(k, v)
        }
        if (restCfg.basicAuth != null) {
            headers.setBasicAuth(restCfg.basicAuth!!.user, restCfg.basicAuth!!.password)
        }
        return headers
    }

    val body: String = requestEnvelope.body

    val method: String = requestEnvelope.method

    companion object {
        private val CONTENT_TYPE = "${MediaType.APPLICATION_JSON};charset=${Charsets.UTF_8}"

        fun emerge(
            validRequest: ValidRequest
        ): Result<RestServiceRequest> {
            val httpRequest = validRequest.httpRequestEnvelope
            val restCfg = validRequest.unvalidatedRequest.restConfiguration.configs[httpRequest.service]!!
            val operationUrl =
                StringSubstitutor(httpRequest.headers).replace(restCfg.operations[httpRequest.operation]!!)
            val uri = restCfg.url + operationUrl
            return Result.runCatching { URI.create(uri) }
                .map {
                    RestServiceRequest(restCfg, httpRequest, operationUrl, it)
                }.recoverCatching {
                    throw RuntimeException("Failed to create RestServiceRequest cause the url {$uri} is not valid", it)
                }
        }
    }

}

