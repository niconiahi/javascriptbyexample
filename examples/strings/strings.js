// "name" is a _string_ which value is "Jose"
const name = "Jose"
console.log("name", name)

// Strings can be concatenated using the [addition operator](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Addition)
const fullName = name + " Martinez"
console.log("fullName", fullName)

// They can also be concatenated using [template literals](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Addition)
const otherName = "Lucas"
const otherFullName = `${name} Garcia`
console.log("otherFullName", otherFullName)

// You can compare if two strings are the same one using the [strict equality operator](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Strict_equality)
const red = "red"
const blue = "blue"
console.log("colors match?", red === blue)

// The comparison between strings is case-senstive
const green = "green"
const greenUppercased = "GREEN"
console.log("green colors match?", green === greenUppercased)

// That's why, when comparing strings, either the [lowercased](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String/toLocaleLowerCase) or [uppercased](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String/toLocaleUpperCase) versions should be compared
console.log(
  "lowercased green colors match?",
  green.toLocaleLowerCase() === greenUppercased.toLocaleLowerCase(),
)

// There are two ways of getting a specific character within a string. Using the [charAt method](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String/charAt)
const animal = "Lion"
console.log("using char at", animal.charAt(0))

// The other one is using the array-like notation
console.log("using array-like notation", animal[0])

// You can verify if a given variable is a _string_ using the [typeof operator](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/typeof)
console.log("typeof animal", typeof animal)
