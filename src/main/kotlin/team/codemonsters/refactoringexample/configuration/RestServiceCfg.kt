package team.codemonsters.refactoringexample.configuration

data class RestServiceCfg(
    var url: String = "",
    var basicAuth: RestBasicAuth? = null,
    var operations: Map<String, String> = emptyMap()
)
