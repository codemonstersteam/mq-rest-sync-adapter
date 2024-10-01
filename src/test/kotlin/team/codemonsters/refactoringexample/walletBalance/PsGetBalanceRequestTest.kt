package team.codemonsters.refactoringexample.walletBalance

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import java.time.Duration
import java.time.OffsetDateTime

internal class PsGetBalanceRequestTest {

    @Test
    fun `should create valid request`() {
        // arrange
        val requestLifetime = 1000L
        val dateFrom = OffsetDateTime.now().minus(Duration.ofSeconds(10))
        val dateTo = OffsetDateTime.now().minus(Duration.ofSeconds(5))
        val operationTime =
            OffsetDateTime.now().plus(
                Duration.ofMillis(60000L),
            ).plus(Duration.ofSeconds(20))

        val jsonString = """
            {
                "requestType": "operationHistory",
                "dateFrom": "$dateFrom",
                "dateTo": "$dateTo",
                "operTime": "$operationTime"
            }
        """.trimIndent()
        // act
        val result = PsGetBalanceRequest.emerge(jsonString, requestLifetime)

        // assert
        assertThat(result.isSuccess).isTrue
    }

    @Test
    fun `should fail when request type is invalid`() {
        // arrange
        val requestLifetime = 1000L
        val operTime = OffsetDateTime.now()
        val jsonString = """
            {
                "requestType": "invalid",
                "dateFrom": "2020-01-01T00:00:00Z",
                "dateTo": "2020-01-02T00:00:00Z",
                "operTime": "2020-01-03T00:00:00Z"
            }
        """.trimIndent()

        // act
        val result = PsGetBalanceRequest.emerge(jsonString, requestLifetime)

        // assert
        assertThat(result.isFailure).isTrue
        assertThat(result.exceptionOrNull()).isInstanceOf(RuntimeException::class.java)
        assertThat(result.exceptionOrNull()?.message).isEqualTo("Ошибка при валидации входных данных. Не верный тип запроса.")
    }

    @Test
    fun `should fail when request operTime is expired`() {
        // arrange
        val requestLifetime = 6000L
        val dateFrom = OffsetDateTime.now().minus(Duration.ofSeconds(10))
        val dateTo = OffsetDateTime.now().minus(Duration.ofSeconds(5))
        val operationTime = OffsetDateTime.now().plus(Duration.ofSeconds(5))

        val jsonString = """
            {
                "requestType": "operationHistory",
                "dateFrom": "$dateFrom",
                "dateTo": "$dateTo",
                "operTime": "$operationTime"
            }
        """.trimIndent()
        // act
        val result = PsGetBalanceRequest.emerge(jsonString, requestLifetime)

        // Then
        assertThat(result.isFailure).isTrue
        assertThat(result.exceptionOrNull()).isInstanceOf(RuntimeException::class.java)
        assertThat(result.exceptionOrNull()?.message).isEqualTo("Истек срок действия запроса")
    }

    @Test
    fun `should fail when request dateFrom is after dateTo`() {
        // arrange
        val requestLifetime = 6000L
        val dateFrom = OffsetDateTime.now().minus(Duration.ofSeconds(5))
        val dateTo = OffsetDateTime.now().minus(Duration.ofSeconds(10))
        val operationTime = OffsetDateTime.now().plus(Duration.ofSeconds(7))

        val jsonString = """
            {
                "requestType": "operationHistory",
                "dateFrom": "$dateFrom",
                "dateTo": "$dateTo",
                "operTime": "$operationTime"
            }
        """.trimIndent()
        // act
        val result = PsGetBalanceRequest.emerge(jsonString, requestLifetime)

        // Then
        assertThat(result.isFailure).isTrue
        assertThat(result.exceptionOrNull()).isInstanceOf(RuntimeException::class.java)
        assertThat(result.exceptionOrNull()?.message).isEqualTo("Ошибка при валидации входных данных. Параметр dateFrom должен быть раньше dateTo")
    }

}