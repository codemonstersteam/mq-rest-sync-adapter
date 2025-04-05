package team.codemonsters.refactoringexample.walletBalance.infrastructure

import org.springframework.http.HttpHeaders
import org.springframework.http.MediaType
import team.codemonsters.refactoringexample.configuration.RestServiceCfg
import team.codemonsters.refactoringexample.walletBalance.HttpRequestEnvelope
import team.codemonsters.refactoringexample.walletBalance.ValidRequest
import java.net.URI

data class RestServiceRequest private constructor(
    private val restCfg: RestServiceCfg,
    val requestEnvelope: HttpRequestEnvelope,
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

    val operationUrl: String = uri.path

    companion object {
        private val CONTENT_TYPE = "${MediaType.APPLICATION_JSON};charset=${Charsets.UTF_8}"

        fun of(
            validRequest: ValidRequest
        ): RestServiceRequest =
            RestServiceRequest(
                validRequest.configuration(),
                validRequest.httpRequestEnvelope,
                validRequest.serviceUri
            )
    }

}

