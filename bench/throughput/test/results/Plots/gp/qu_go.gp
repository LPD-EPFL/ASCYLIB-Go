load 'gp/style.gp'
set macros
NOYTICS = "set format y ''; unset ylabel"
YTICS = "set ylabel 'Throughput (Mops/s)' offset 2"
PSIZE = "set size 0.3, 0.5"

set key horiz maxrows 1

set output "pdf/qu_go.pdf"

set terminal pdf enhanced size 10,5
set rmargin 0
set lmargin 3
set tmargin 3
set bmargin 2.5

title_offset   = -0.5
ytics_offset   = 0.65
top_row_y      = 0.0
bottom_row_y   = 0.0
graphs_x_offs  = 0.03
graphs_y_offs  = 0.12
plot_size_x    = 1
plot_size_y    = 1.5

DIV              =    1e6
FIRST            =    2
OFFSET           =    3
column_select(i) = column(FIRST + (i*OFFSET)) / (DIV);

LINE0 = '"queue-ms-lb"'
LINE1 = '"queue-ms-lf"'
LINE2 = '"queue-optik1"'
LINE3 = '"queue-optik2"'
LINE4 = '"priorityqueue-lotanshavit-lf"'

PLOT0 = '"Decreasing size\n{/*0.8(40% enqueue, 60% dequeue)}"'
PLOT1 = '"Stable size\n{/*0.8(50% enqueue, 50% dequeue)}"'
PLOT2 = '"Increasing size\n{/*0.8(60% enqueue, 40% dequeue)}"'

# ##########################################################################################
# XEON #####################################################################################
# ##########################################################################################

FILE0 = '"data/data.qu.thr.p40.dat"'
FILE1 = '"data/data.qu.thr.p50.dat"'
FILE2 = '"data/data.qu.thr.p60.dat"'

set xlabel "# Threads" offset 0, 0.75 font ",14"
set xtics offset 0,0.4
unset key


set size plot_size_x, plot_size_y
set multiplot layout 5, 2

set origin 0.0 + graphs_x_offs, top_row_y + 1 * (0.38 + graphs_y_offs)
@PSIZE
set title @PLOT0 offset 0.2,title_offset font ",14"
set ylabel 'Throughput (Mops/s)' offset 1,0.5
set ytics offset ytics_offset
plot \
     @FILE0 using 1:(column_select(0)) title @LINE0 ls 1 with linespoints, \
     "" using 1:(column_select(1)) title @LINE1 ls 2 with linespoints, \
     "" using 1:(column_select(2)) title @LINE2 ls 3 with linespoints, \
     "" using 1:(column_select(3)) title @LINE3 ls 4 with linespoints, \
     "" using 1:(column_select(4)) title @LINE4 ls 5 with linespoints

set origin 0.0 + graphs_x_offs, top_row_y
@PSIZE
set title @PLOT1
set ylabel 'Throughput (Mops/s)' offset 1,0.5
set ytics offset ytics_offset
plot \
     @FILE1 using 1:(column_select(0)) title @LINE0 ls 1 with linespoints, \
     "" using 1:(column_select(1)) title @LINE1 ls 2 with linespoints, \
     "" using 1:(column_select(2)) title @LINE2 ls 3 with linespoints, \
     "" using 1:(column_select(3)) title @LINE3 ls 4 with linespoints, \
     "" using 1:(column_select(4)) title @LINE4 ls 5 with linespoints

set origin 0.325 + graphs_x_offs, top_row_y + 1 * (0.38 + graphs_y_offs)
@PSIZE
set title @PLOT2
@YTICS
set ylabel ""
unset ylabel
plot \
     @FILE2 using 1:(column_select(0)) title @LINE0 ls 1 with linespoints, \
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
set key at screen 0.4,screen 0.3 left top

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
