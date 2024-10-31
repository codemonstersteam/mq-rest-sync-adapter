package team.codemonsters.refactoringexample.walletBalance.application

import org.slf4j.LoggerFactory
import org.springframework.http.HttpMethod
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Component
import org.springframework.web.reactive.function.client.ClientResponse
import org.springframework.web.reactive.function.client.WebClient
import reactor.core.publisher.Mono
import reactor.kotlin.core.publisher.toMono
import team.codemonsters.refactoringexample.walletBalance.infrastructure.RestServiceRequest
import team.codemonsters.refactoringexample.walletBalance.infrastructure.RestServiceResponse

@Component
class RestGateway(
    private val webClient: WebClient
) {
    private val log = LoggerFactory.getLogger(WalletBalanceService::class.java)!!

    fun sendRequest(request: RestServiceRequest): Mono<Result<RestServiceResponse>> {
        log.info("sending RestServiceRequest ${request}")
        return when (request.method) {
            HttpMethod.POST.name() -> post(request)
            HttpMethod.GET.name() -> get(request)
            else -> RestServiceResponse.unsupportedMethod(request).toMono()
        }.map {
            Result.success(it)
        }
    }

    private fun get(
        request: RestServiceRequest
    ): Mono<RestServiceResponse> {
        return webClient.get()
            .uri(request.uri)
            .headers { request.headers() }
            .exchangeToMono { response ->
                response.bodyToMono(String::class.java)
                    .switchIfEmpty("".toMono())
                    .map { mapToSuccessResponse(request, response, it) }
            }
            .onErrorResume { RestServiceResponse.error(request, it).toMono() }
    }

    private fun post(
        request: RestServiceRequest
    ) =
        webClient.post()
            .uri(request.uri)
            .headers { request.headers() }
            .bodyValue(request.body)
            .exchangeToMono { response ->
                response.bodyToMono(String::class.java)
                    .switchIfEmpty("".toMono())
                    .map { mapToSuccessResponse(request, response, it) }
            }
            .onErrorResume { error -> Mono.just(RestServiceResponse.error(request, error)) }

    private fun mapToSuccessResponse(
        request: RestServiceRequest,
        response: ClientResponse,
        body: String
    ) =
        RestServiceResponse.success(
            request,
            HttpStatus.valueOf(response.statusCode().value()),
            response.headers().asHttpHeaders(),
            body
        )

}