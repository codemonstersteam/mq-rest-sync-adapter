package team.codemonsters.refactoringexample.walletBalance

data class WalletBalance(val validRequest: ValidRequest, val restServiceResponse: RestServiceResponse) {
    val httpResponseEnvelope =
        HttpResponseEnvelope.of(restServiceResponse)
}