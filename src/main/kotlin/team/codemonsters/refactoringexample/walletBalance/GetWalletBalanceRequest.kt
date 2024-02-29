package team.codemonsters.refactoringexample.walletBalance

import java.time.OffsetDateTime


data class GetWalletBalanceRequest(
    val requestType: String,
    val dateFrom: OffsetDateTime,
    val dateTo: OffsetDateTime?,
    val additionalParameters: Map<String, String>
) {
    companion object {
        fun of(
            request: PsGetBalanceRequest
        ): GetWalletBalanceRequest =
            GetWalletBalanceRequest(
                request.requestType,
                request.dateFrom,
                request.dateTo,
                mapOf()
            )
    }
}