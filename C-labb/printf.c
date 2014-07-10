/*
	Author : Christopher Lillthors.
	CDATE 1, KTH.
*/

#include <stdio.h>
//Various examples of printf.
int main(int argc, char const *argv[])
{
	long double pi = 3.14159;
	printf("The value of pi: %Lf\n", pi);
	printf("The value of pi with 2 decimals: %.2Lf\n",pi);
	printf("The value of pi written with short notation: %Lg\n", pi);
	printf("The memory address of pi: %p\n",&pi);

	char name[] = "Kalle";
	printf("%s\n", name);
	printf("%p\n", name); //Memory address to the first char.
	/*
	0 1 2 3 4
	K a  l  l e
	*/
	return 0; //close program, and return success.
}