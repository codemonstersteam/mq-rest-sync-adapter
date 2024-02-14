package team.codemonsters.refactoringexample.walletBalance

import team.codemonsters.refactoringexample.exceptions.ServiceCommunicationException

import java.time.Duration
import java.time.OffsetDateTime

data class ValidRequest private constructor(
    val httpRequestEnvelope: HttpRequestEnvelope
){
    companion object {
        fun emerge(httpRequestEnvelope: HttpRequestEnvelope): Result<ValidRequest> {

            PsGetBalanceRequest.fromJsonString(httpRequestEnvelope.body!!)
                .map{

                    if (OffsetDateTime.now().plus(Duration.ofMillis(timeout)).isAfter(it.operTime)) {
                        return Result.failure(
                            ServiceCommunicationException("validation", "Истек скорок действия запроса")
                        )
                    }
                    if (OperationType.OPERATION_HISTORY == getOperationType(it.requestType)
                        && it.dateFrom.isBefore(it.dateTo ?: OffsetDateTime.now())
                        && (it.dateTo == null || it.dateTo.isBefore(OffsetDateTime.now()))
                    ) {
                        return Result.success(data)
                    } else {
                        return Result.failure(
                            ServiceCommunicationException("validation", "Ошибка при валидации входных данных")
                        )
                    }
                }
            )

        }
    }

}