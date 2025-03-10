// TypeScript sample file

interface TestInterface {
  value: string;
  test(): boolean;
}

function testFunction<T>(input: T): T {
  console.log('Testing with input:', input);
  return input;
}

class TestImplementation implements TestInterface {
  constructor(public value: string) {}
  
  test(): boolean {
    return this.value.toLowerCase().includes('test');
  }
  
  private testHelper(): string {
    return `The value is: ${this.value}`;
  }
}

// Generic test class
class GenericTest<T> {
  constructor(private items: T[]) {}
  
  test(predicate: (item: T) => boolean): T[] {
    return this.items.filter(predicate);
  }
}

// Example usage
const testObj = new TestImplementation('This is a test');
console.log(testObj.test());  // Should return true

const numbers = new GenericTest<number>([1, 2, 3, 4, 5]);
const evenNumbers = numbers.test(num => num % 2 === 0);
console.log('Even numbers:', evenNumbers);

// Using the test function
const result = testFunction<string>('Hello test world');
console.log(result);