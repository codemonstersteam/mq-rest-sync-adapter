package team.codemonsters.refactoringexample.walletBalance

import java.time.OffsetDateTime


data class GetWalletBalanceRequest(
    val requestType: String,
    //@JsonFormat(shape = JsonFormat.Shape.STRING, pattern = "yyyy-MM-dd'T'HH:mm:ss'Z'")
    val dateFrom: OffsetDateTime,
    //@JsonFormat(shape = JsonFormat.Shape.STRING, pattern = "yyyy-MM-dd'T'HH:mm:ss'Z'")
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