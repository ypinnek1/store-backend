Array.prototype.concat = function (...args) {
    let results = []
    let initialArr = this
    for (let i = 0; i < initialArr.length; i++) {
        results.push(initialArr[i])
    }

    for (let j = 0; j < args.length; j++) {
        if (Array.isArray(args[j])) {
            let arr = args[j]
            for (let k = 0; k < arr.length; k++) {
                results.push(arr[k])
            }
        } else {
            results.push(args[j])
        }
    }
    return results
}

let arr = [10, 20, 30]
let newArr = arr.concat(40, [50, 60], [70, [80, 90]])
// console.log(newArr)

//[10, 20, 30, 40, 50, 60, 70, [80, 90]]


Array.prototype.slice = function (args) {
    let results = [];
    console.log("args = ", args)
}

const animals = ['ant', 'bison', 'camel', 'duck', 'elephant'];

console.log(animals.slice(2, 4));