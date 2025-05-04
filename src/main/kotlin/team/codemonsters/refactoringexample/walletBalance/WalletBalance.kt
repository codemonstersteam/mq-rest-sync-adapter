package team.codemonsters.refactoringexample.walletBalance

import team.codemonsters.refactoringexample.walletBalance.infrastructure.RestServiceResponse

data class WalletBalance(val validRequest: ValidRequest, val restServiceResponse: RestServiceResponse) {
    val httpResponseEnvelope =
        HttpResponseEnvelope.of(restServiceResponse)
}