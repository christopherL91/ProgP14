/*
    Original work:
    Unknown.
    Modified by:
    Christopher Lillthors.
*/


#include <stdio.h>
#include <stdlib.h>

typedef struct post {
    char name[29];
    float bmi;
    struct post * next;
}Post;

void writePost(Post *p); //function declaration.

int main(int argc, char * argv[]) {
    int weight; //create weight variable
    float length; //create length variable
    Post * p = (Post *)malloc(sizeof(Post)); //create new empty struct 
    p -> next = (Post *)malloc(sizeof(Post)); 
    /*
        Create new empty struct, and let
        next point to it.
    */

    printf("Vad heter du? ");
    scanf("%s", p->name);
    /*read string from stdin*/

    printf("Hur lång är du (m)? ");
    scanf("%f", &length);
    /*read float from stdin*/

    printf("Vad väger du (kg)? ");
    scanf("%d", &weight);
    /*read in int from stdin*/

    /*Calculate bmi and and assign p.bmi to it*/
    p -> bmi = weight / (length * length);

    writePost(p); /*Send pointer to struct, and execute function*/

    free(p); //Free memory. 
    free(p->next); //Free memory. 

    return 0; //Standard in UNIX.
}

void writePost(Post * p) {
    printf("Namn: %s\nbmi: %.2f\n", p->name, p->bmi); 
}
