package team.codemonsters.refactoringexample.mq

import com.fasterxml.jackson.annotation.JsonFormat
import team.codemonsters.refactoringexample.util.JsonUtil
import java.time.OffsetDateTime


data class GetWalletBalanceRequest(
    val requestType: String,
    @JsonFormat(shape = JsonFormat.Shape.STRING, pattern = "yyyy-MM-dd'T'HH:mm:ss'Z'")
    val dateFrom: OffsetDateTime,
    @JsonFormat(shape = JsonFormat.Shape.STRING, pattern = "yyyy-MM-dd'T'HH:mm:ss'Z'")
    val dateTo: OffsetDateTime?,
    val additionalParameters: AdditionalParameters
) {

    companion object {
        fun emerge(
            data: String
        ): GetWalletBalanceRequest {
            val request = JsonUtil.fromString(data, PsGetBalanceRequest::class.java)
            return createRequest(request)
        }

        private fun createRequest(
            request: PsGetBalanceRequest
        ): GetWalletBalanceRequest =
            GetWalletBalanceRequest(
                request.requestType,
                request.dateFrom,
                request.dateTo,
                AdditionalParameters()
            )
    }
}