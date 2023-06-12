package team.codemonsters.refactoringexample.exceptions

class ProcessException: RuntimeException {

    val code: String
    val isBusinessError: Boolean

    constructor(code: String, isBusinessError: Boolean) {
        this.code = code
        this.isBusinessError = isBusinessError
    }

    constructor(code: String, e: Throwable?) : super(e?.message, e) {
        this.code = code
        this.isBusinessError = false
    }
}