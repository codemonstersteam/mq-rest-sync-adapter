package team.codemonsters.refactoringexample.valueObject

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

internal class EmailTest {

    @Test
    fun `should not create email with invalid email`() {
        // arrange
        val invalidEmail = "invalidEmail"
        // act
        val result = Email.emerge(invalidEmail)
        // assert
        assertThat(result.isFailure).isTrue
    }

    @Test
    fun `should create email with valid email`() {
        // arrange
        val validEmail = "hPbqJ@example.com"
        // act
        val result = Email.emerge(validEmail)
        // assert
        assertThat(result.isSuccess).isTrue
    }
}