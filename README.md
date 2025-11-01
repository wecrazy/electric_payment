# electric_payment
Prototype of Electric Payment via Website for PLTMH Lembang Palesan

```                                                                             
                      xxxxxxxxxxxx:  xxxxx.   xxxxxxxxxx                      
                     +x.        ;x;  xx     ;xx;       xx                     
                     xx  +xxxx: ;x;  .     +;xx; ;xx+  xx                     
                     xx  +xxxx: ;x;      ;++ xx; ;xx+  xx                     
                     xx  ;xxxx. ;x;  x; +++. xx;       xx                     
                     xx         ;x;   :+++;  xxxxxxxxxxxx                     
                     ;xxxxxxxxxxxx   ++++x                                    
                                   .+++x+;.......   ;;;;.                     
                     ;++.  ;+++;  ++++++++++++++    +++xx                     
                     xx.    ... .+++x+++x++++x;        xx                     
                     xx               .+++x++.   .xxxxxx+                     
                     xxxxxxxxx;       ;+++++                                  
                     xx          xx  :++x+.  xxxx  ;xxxxx                     
                     xx  +xxxxx  xx  ++++    xx.       xx                     
                     xx  +xxxxx  x; ;++.  xx xx. ;xxx  xx                     
                     xx  +xxxxx    .++    xx xx. ;xx+  xx                     
                     +x.           +.    :xx xx.       xx                     
                      +xxxxxxxxxx     xxxxx+ +xxxxxxxxx+                      
                                                                              
                                                                              
              .$$$$. $.    $$$$+  :$$X. X$$$$X X$$X;  Xx   x$$;               
              :$     $;    $.    $X       $$   $x  $x $X X$   .               
              :$$XX  $;    $$XX ;$        $$   $$$$$  $X $x                   
              :$;..  $x..  $+... X$;.x$   $$   $x X$  $X .$x.:$.              
                                    .                       .                 
                              ;;;.    ;:  ;:   ;.                             
                             .$. +$  X$$.  $+ $+                              
                             .$xx$$ .$ :$   $$;                               
                             .$     $$XX$X  ;$                                
                              :     .    :   .                                
                                                                              
```                                                                                                                                
Build with ❤ using Golang v1.25.0

***PODMAN***
1) podman build -t electric-payment:prod -f ContainerFile .
2) podman save -o electric-payment-prod.tar localhost/electric-payment:prod