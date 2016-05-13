#!/usr/bin/awk -f
#
# run like: ps xu | ./ps.awk
# or like: watch -n0.2 'ps xu | ./ps.awk'

BEGIN {
  print("   CPU\t     MEM\tPID\t\tPROCESS")
  cpu=0
  mem=0
} /metrics-capacitor \(/ {
  mem+=$6
  cpu+=$3
  printf "%5s%\t%6sMB\t%s\t\t%s\n", $3, sprintf("%4.1f",$6/1024), $2, $11" "$12
} END {
  print "-----------------------------------------------------------------------"
  printf "%5s%\t%6sMB\t--\t\tALL\n", cpu, sprintf("%4.1f", mem/1024)
}
