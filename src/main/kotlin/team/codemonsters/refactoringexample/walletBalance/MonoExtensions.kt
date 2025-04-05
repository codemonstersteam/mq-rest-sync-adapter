package team.codemonsters.refactoringexample.walletBalance

import reactor.core.publisher.Mono
import reactor.kotlin.core.publisher.toMono

fun <T, R> Mono<Result<T>>.pipe(function: (T) -> Mono<Result<R>>)
        : Mono<Result<R>> =
    this.flatMap { result ->
        result.fold(
            { value -> function(value) },
            { exception -> Result.failure<R>(exception).toMono() }
        )
    }

fun <T, R> Mono<Result<T>>.fold(function: (T) -> R)
        : Mono<Result<R>> =
    this.map { result ->
        result.fold(
            { value -> Result.success(function(value)) },
            { exception -> Result.failure(exception) }
        )
    }