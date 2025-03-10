// This is a JavaScript sample file

function test(input) {
  // Test the input
  if (typeof input !== 'string') {
    throw new Error('Input must be a string');
  }
  
  return input.toUpperCase();
}

class TestClass {
  constructor(value) {
    this.value = value;
  }
  
  test() {
    console.log('Testing value:', this.value);
    return this.value * 2;
  }
  
  static testStatic() {
    return 'This is a static test method';
  }
}

// Some example usage
const result = test('Hello world');
console.log(result);

const testObj = new TestClass(42);
console.log(testObj.test());
console.log(TestClass.testStatic());