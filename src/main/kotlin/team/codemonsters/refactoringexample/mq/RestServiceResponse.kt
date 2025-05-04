package team.codemonsters.refactoringexample.mq

import org.springframework.http.HttpHeaders
import org.springframework.http.HttpStatus


data class RestServiceResponse (
    val serviceBaseUrl: String,
    val requestEnvelope: HttpRequestEnvelope,
    val httpCode: HttpStatus,
    val headers: HttpHeaders,
    val body: String?){

    companion object {
        fun emerge(
            request: RestServiceRequest,
            httpStatus: HttpStatus,
            headers: HttpHeaders,
            body: String?
        ): Result<RestServiceResponse> {
            return Result.success(RestServiceResponse(request.restCfg.url, request.requestEnvelope, httpStatus, headers, body))
        }
    }
}