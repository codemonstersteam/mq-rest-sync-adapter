package team.codemonsters.refactoringexample.exceptions

class ServiceCommunicationException (serviceName: String, msg: String):
    RuntimeException("Ошибка получения данных с сервиса '%s': %s".format(serviceName, msg))