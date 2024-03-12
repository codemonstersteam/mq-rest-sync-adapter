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
            HttpRequestEnvelopeDto.parseJsonString(
                unvalidatedRequest.payload.toString()
            )
                .flatMap {
                    HttpRequestEnvelope.emerge(
                        it,
                        unvalidatedRequest.restConfiguration
                    )
                }
                .flatMap { validRestRequest(unvalidatedRequest, it) }

        private fun validRestRequest(
            unvalidatedRequest: UnvalidatedRequest,
            httpRequestEnvelope: HttpRequestEnvelope,
        ) =
            PsGetBalanceRequest.emerge(
                httpRequestEnvelope.body,
                unvalidatedRequest.requestLifetime
            )
                .map {
                    ValidRequest(
                        unvalidatedRequest,
                        it,
                       httpRequestEnvelope,
                    )
                }
    }

}
