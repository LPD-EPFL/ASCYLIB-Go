load 'gp/style.gp'
set macros
NOYTICS = "set format y ''; unset ylabel"
YTICS = "set ylabel 'Throughput (Mops/s)' offset 2"
PSIZE = "set size 0.45, 0.95"

set key horiz maxrows 1

set output 'eps/qu_remove_'.hostname.'.eps'

set terminal eps enhanced size 6.5,2.5
set rmargin 0
set lmargin 3
set tmargin 3
set bmargin 2.5

title_offset   = -0.5
ytics_offset   = 0.65
top_row_y      = 0.0
bottom_row_y   = 0.0
graphs_x_offs  = 0.055
graphs_y_offs  = 0.12
plot_size_x    = 1
plot_size_y    = 1.5

col_get = 0
col_set = 1
col_rem = 2

FIRST            = 2
OFFSET           = 3
column_select(i) = column(FIRST + (i * OFFSET) + col_rem);

LINE0 = '"ms-lb"'
LINE1 = '"ms-lf"'
LINE2 = '"optik1"'
LINE3 = '"optik2"'
LINE4 = '"prio-lotanshavit-lf"'

PLOT0 = '"Only contention\n{/*0.6(remove ops, 100% updates)}"'

# ##########################################################################################
# XEON #####################################################################################
# ##########################################################################################

FILE0 = '"data/'.hostname.'.qu.u100.dat"'

set xlabel "# Threads" offset 0, 0.75 font ",14"
set xtics offset 0,0.4
unset key


set size plot_size_x, plot_size_y
set multiplot layout 5, 2

set origin 0.055, 0.035
@PSIZE
set title @PLOT0 offset 0.2,title_offset font ",14"
set ylabel 'Latency (µs/ops)' offset 1.5,0.5
set ytics offset ytics_offset
plot \
     @FILE0 using 1:(column_select(0)) title @LINE0 ls 1 with linespoints, \
     "" using 1:(column_select(1)) title @LINE1 ls 2 with linespoints, \
     "" using 1:(column_select(2)) title @LINE2 ls 3 with linespoints, \
     "" using 1:(column_select(3)) title @LINE3 ls 4 with linespoints, \
     "" using 1:(column_select(4)) title @LINE4 ls 5 with linespoints

unset origin
unset border
unset tics
unset xlabel
unset label
unset arrow
unset title
unset object

#Now set the size of this plot to something BIG
set size plot_size_x, plot_size_y #however big you need it
set origin 0.0, 1.1

#Key settings
set key vertical Left samplen 4 maxrows 10 maxcols 2
set key at screen 0.6,screen 0.65 left top

#We need to set an explicit xrange.  Anything will work really.
set xrange [-1:1]
@NOYTICS
set yrange [-1:1]
plot \
     NaN title @LINE0 ls 1 with linespoints, \
     NaN title @LINE1 ls 2 with linespoints, \
     NaN title @LINE2 ls 3 with linespoints, \
     NaN title @LINE3 ls 4 with linespoints, \
     NaN title @LINE4 ls 5 with linespoints

#</null>
unset multiplot  #<--- Necessary for some terminals, but not postscript I don't thin
