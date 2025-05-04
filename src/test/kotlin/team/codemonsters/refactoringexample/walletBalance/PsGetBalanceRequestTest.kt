package team.codemonsters.refactoringexample.walletBalance

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import java.time.Duration
import java.time.OffsetDateTime

internal class PsGetBalanceRequestTest {

    @Test
    fun `should create valid request`() {
        // arrange
        val requestLifetime = Duration.ofSeconds(10).toMillis()
        val dateFrom = OffsetDateTime.now().minus(Duration.ofDays(10))
        val dateTo = OffsetDateTime.now()
        val operationTime = OffsetDateTime.now().minus(Duration.ofMillis(requestLifetime))

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
        val requestLifetime = Duration.ofSeconds(10).toMillis()
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
        assertThat(result.exceptionOrNull()?.message).isEqualTo("Ошибка при валидации входных данных. Не верный тип запроса.")
    }

    @Test
    fun `should fail when jsonString is invalid JSON`() {
        val jsonString = "invalid json"
        val requestLifetime = Duration.ofHours(1).toMillis()

        val result = PsGetBalanceRequest.emerge(jsonString, requestLifetime)

        assertThat(result.exceptionOrNull()?.message).contains("Failed to parse JSON")
    }

    @Test
    fun `should fail when request operTime is expired`() {
        // arrange
        val requestLifetime = Duration.ofSeconds(10).toMillis()
        val dateFrom = OffsetDateTime.now().minus(Duration.ofSeconds(10))
        val dateTo = OffsetDateTime.now().minus(Duration.ofSeconds(5))
        val operationTime = OffsetDateTime.now()

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
        assertThat(result.exceptionOrNull()?.message).isEqualTo("Истек срок действия запроса")
    }

    @Test
    fun `should fail when request dateFrom is after dateTo`() {
        // arrange
        val requestLifetime = Duration.ofSeconds(10).toMillis()
        val dateFrom = OffsetDateTime.now().minus(Duration.ofSeconds(5))
        val dateTo = OffsetDateTime.now().minus(Duration.ofSeconds(10))
        val operationTime = OffsetDateTime.now().minus(Duration.ofMillis(requestLifetime))

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
        assertThat(result.exceptionOrNull()?.message).isEqualTo("Ошибка при валидации входных данных. Параметр dateFrom должен быть раньше dateTo")
    }

}