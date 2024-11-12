#include <stdio.h>
#include "utils.h"

// Example function that may violate MISRA rules
int example_function(int x) {
    int uninitialized_variable;
    if (x > 0) {
        uninitialized_variable = x * 2;
    }
    return uninitialized_variable; // Potential MISRA violation: use of uninitialized variable
}

int main() {
    int result = add(3, 4);
    printf("Result: %d\n", result);

    int example_result = example_function(5);
    printf("Example Result: %d\n", example_result);
    
    return 0;
}
