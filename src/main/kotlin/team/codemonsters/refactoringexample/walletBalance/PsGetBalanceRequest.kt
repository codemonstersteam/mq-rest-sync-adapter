package team.codemonsters.refactoringexample.walletBalance

import team.codemonsters.refactoringexample.util.JsonUtil
import java.time.Duration
import java.time.OffsetDateTime

data class PsGetBalanceRequest(
    val requestType: String,
    val dateFrom: OffsetDateTime,
    val dateTo: OffsetDateTime?,
    val operTime: OffsetDateTime
) {
    companion object {
        fun emerge(
            jsonString: String,
            requestLifetime: Long,
        ): Result<PsGetBalanceRequest> =
            parseJsonString(jsonString)
                .ensure(
                    { "operationHistory" == it.requestType },
                    "Ошибка при валидации входных данных. Не верный тип запроса.",
                )
                .ensure(
                    {
                        OffsetDateTime.now().plus(Duration.ofMillis(requestLifetime))
                            .isBefore(it.operTime)
                    },
                    "Истек срок действия запроса",
                )
                .ensure(
                    { it.dateFrom.isBefore(it.dateTo ?: OffsetDateTime.now()) },
                    "Ошибка при валидации входных данных. Параметр dateFrom должен быть раньше dateTo",
                )

        private fun parseJsonString(jsonString: String) =
            Result.runCatching { JsonUtil.fromString(jsonString, PsGetBalanceRequest::class.java) }
    }
}
