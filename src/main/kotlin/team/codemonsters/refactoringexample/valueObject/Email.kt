package team.codemonsters.refactoringexample.valueObject

val EMAIL_ADDRESS = Regex(
    "[a-zA-Z0-9\\+\\.\\_\\%\\-\\+]{1,256}" +
            "\\@" +
            "[a-zA-Z0-9][a-zA-Z0-9\\-]{0,64}" +
            "(" +
            "\\." +
            "[a-zA-Z0-9][a-zA-Z0-9\\-]{0,25}" +
            ")+"
)

data class Email private constructor(
    val email: String
) {
    companion object {
        fun emerge(email: String): Result<Email> =
            if (EMAIL_ADDRESS
                    .matches(email)
            )
                Result.success(Email(email))
            else
                Result.failure(IllegalArgumentException("Invalid email: $email"))

    }
}
