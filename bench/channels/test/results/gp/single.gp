load 'gp/style.gp'
set macros
NOYTICS = "set format y ''; unset ylabel"
YTICS = "set ylabel 'Throughput (Mops/s)' offset 2"
PSIZE = "set size 0.42, 0.5"

set key horiz maxrows 1

set output 'pdf/single.'.hostname.'.'.mode.'.s'.servers.'.pdf'

set terminal pdf enhanced size 6.5,5
set rmargin 0
set lmargin 3
set tmargin 3
set bmargin 2.5

title_offset   = -0.5
ytics_offset   = 0.65
top_row_y      = 0.0
bottom_row_y   = 0.0
graphs_x_offs  = 0.07
graphs_y_offs  = 0.12
plot_size_x    = 1
plot_size_y    = 1.5

col_msgs = 0
col_tput = 1
col_lat  = 2
column_select(i) = column(2 + i);

LINE0 = '"Exchanges"'
LINE1 = '"Throughput"'
LINE2 = '"Latency"'

PLOT0 = '"Exchanges\n{/*0.6('.hostname.', '.servers.' server(s), '.mode.' mode)}"'
PLOT1 = '"Throughput\n{/*0.6('.hostname.', '.servers.' server(s), '.mode.' mode)}"'
PLOT2 = '"Latency\n{/*0.6('.hostname.', '.servers.' server(s), '.mode.' mode)}"'

FILE = '"data/'.hostname.'.'.mode.'.s'.servers.'.dat"'

set xlabel "# Clients" offset 0, 0.75 font ",14"
set xtics offset 0,0.4
unset key

set size plot_size_x, plot_size_y
set multiplot layout 2, 2

set origin 0.0 + graphs_x_offs, top_row_y + 1 * (0.38 + graphs_y_offs)
@PSIZE
set title @PLOT0 offset 0.2,title_offset font ",14"
set ylabel 'Exchanges (msg×10^6)' offset 1.5,0.5
set ytics offset ytics_offset
plot \
     @FILE using 1:(column_select(col_msgs))/1e6 title @LINE0 ls 1 with linespoints

set origin 0.0 + graphs_x_offs, top_row_y
@PSIZE
set title @PLOT1
set ylabel 'Throughput (MB/s)' offset 1.5,0.5
set ytics offset ytics_offset
plot \
     @FILE using 1:(column_select(col_tput)) title @LINE1 ls 2 with linespoints

set origin 0.475 + graphs_x_offs, top_row_y + 1 * (0.38 + graphs_y_offs)
@PSIZE
set title @PLOT2
set ylabel 'Latency (µs/msg)' offset 1.5,0.5
set ytics offset ytics_offset
plot \
     @FILE using 1:(column_select(col_lat)) title @LINE2 ls 3 with linespoints

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
set key at screen 0.67,screen 0.3 left top

#We need to set an explicit xrange.  Anything will work really.
set xrange [-1:1]
@NOYTICS
set yrange [-1:1]
plot \
     NaN title @LINE0 ls 1 with linespoints, \
     NaN title @LINE1 ls 2 with linespoints, \
     NaN title @LINE2 ls 3 with linespoints

#</null>
unset multiplot  #<--- Necessary for some terminals, but not postscript I don't thin
