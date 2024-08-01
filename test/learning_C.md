# Learning C 


## includes
You can include external functions by adding includes
```c filename=main.c
#include <stdio.h>
#include <stdlib.h>
```



## Structs

This is how you define a Struct

```c filename=definitions.h
typedef struct YourMom {
    int timesPounded;
} YourMom_T;

int add(int a, int b);
```

## Main 

```c  filename=main.c
#include "definitions.h"


int main(){
    YourMom_T* your_mom = (YourMom_T*)malloc(sizeof(YourMom_T));
    your_mom->timesPounded = 69;
    printf("hello world\n");
    printf("%d\n", your_mom->timesPounded);
    printf("%d\n", add(1,2));
}

int add(int a, int b){
    return a + b;
}
```

