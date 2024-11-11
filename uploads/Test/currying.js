function curry(func) {
  // The number of arguments the original function expects
  const arity = func.length;
  // Inner function to collect arguments
  function curried(...args) {
    console.log("args = ", ...args);
    // If we have enough arguments, invoke the original function
    if (args.length >= arity) {
      return func(...args);
    } else {
      // Otherwise, return a function that will collect more arguments
      return (...nextArgs) => curried(...args, ...nextArgs);
    }
  }

  return curried;
}

// Test examples
function add(a, b) {
  return a + b;
}

const curriedAdd = curry(add);
console.log(curriedAdd(3)(4)); // Output: 7

const alreadyAddedThree = curriedAdd(3);
console.log("alreadyAddedThree = ", alreadyAddedThree);
console.log(alreadyAddedThree(4)); // Output: 7

// function multiplyThreeNumbers(a, b, c) {
//   return a * b * c;
// }

// const curriedMultiplyThreeNumbers = curry(multiplyThreeNumbers);
// console.log(curriedMultiplyThreeNumbers(4)(5)(6)); // Output: 120

// const containsFour = curriedMultiplyThreeNumbers(4);
// const containsFourMulFive = containsFour(5);
// console.log(containsFourMulFive(6)); // Output: 120
