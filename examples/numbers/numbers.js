// _numbers_ are numbers

// You can sum _numbers_
const two = 2
const four = 4
const summed = four + two
console.log("four plus two is", summed)

// You can divide _numbers_
const divided = four / two
console.log("four divided by two is", divided)

// You can substract _numbers_
const substracted = four - two
console.log("for minus two is", substracted)

// You can multiply _numbers_
const multiplied = four * two
console.log("for times two is", multiplied)

// You can get the remainder of one _number_ over another with the [remainder operator](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Remainder)
const remainder = four % two
console.log("the remainder is", remainder)

// It's possible to [typecast](https://en.wikipedia.org/wiki/Type_conversion) a string to a _number_
const casted = Number("2")
console.log(
  `casting the string "2" to number results in`,
  casted,
)

// If the type cast fails, then [NaN](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/NaN) is returned
const castedFail = Number("Sun")
console.log(
  `casting the string "Sun" to number results in`,
  castedFail,
)

// You can verify if a given variable is a _number_ using the [typeof operator](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/typeof)
console.log(`the type of "two" is`, typeof two)
