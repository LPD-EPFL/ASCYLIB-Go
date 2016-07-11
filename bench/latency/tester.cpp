/**
 * @file   tester.cpp
 * @author Sébastien Rouault <sebastien.rouault@epfl.ch>
 *
 * @section LICENSE
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 3 of the License, or
 * any later version. Please see https://gnu.org/licenses/gpl.html
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * @section DESCRIPTION
 *
 * Latency measurement program.
**/

// ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
// ▁ Declarations ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
// ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔

// External headers
#include <cstddef>
#include <fstream>
#include <iostream>
extern "C" {
    #include <sys/types.h>
    #include <sys/utsname.h>
    #include <sys/wait.h>
    #include <unistd.h>
}

// ―――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

// Null constant
#define null nullptr

// More types
typedef uint_fast32_t nat_t;
typedef double        val_t;
typedef int           pid_t;
typedef int           fd_t;

// ―――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

// Constants/configuration
namespace Constants {
    constexpr nat_t reps = 5; // Amount of repetitions
    constexpr nat_t load[] = { 0, 20, 50, 100 }; // Loads tested
    constexpr nat_t load_length = sizeof(load) / sizeof(nat_t); // Number of elements in table 'load'
    constexpr nat_t core_divs = 32; // Number of divisions
    static_assert(Constants::core_divs >= 1, "Not enough divisions (at least 1)");
}

// ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
// ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔ Declarations ▔
// ▁ Tools ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
// ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔

/** Numeric to string converter.
 * @param size Size of the string
**/
template<nat_t size = 16> class Converter final {
private:
    char data[size];
public:
    /** Constructor.
     * @param num Number to convert and hold
    **/
    Converter(nat_t num) {
        nat_t i = 0;
        { // Computing limit value of i
            nat_t tmp = num;
            do {
                i++;
                tmp /= 10;
            } while (tmp != 0);
        }
        if (i >= size) // Number too large
            std::terminate();
        data[i] = '\0';
        do {
            data[--i] = (num % 10) + '0';
            num /= 10;
        } while (num != 0);
    }
    /** Implicit conversion to pointer to char.
     * @return Pointer to char holding the number
    **/
    operator char*() {
        return data;
    }
};

// ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
// ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔ Tools ▔
// ▁ Test ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
// ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔

/** Test for a given configuration.
**/
class Test final {
private:
    /** Simple on-line float parser.
    **/
    class Parser final {
    private:
        val_t current; // Current value
        nat_t pt_pos;  // Position after the point (0 for before)
        bool  is_nan;  // True if not a number, false otherwise
    public:
        /** Constructor.
        **/
        Parser(): current(0), pt_pos(0), is_nan(false) {
        }
    public:
        /** Push a table of characters, ended by a new line '\n'.
         * @param buf Buffer to push
         * @param len Buffer length
         * @return Position of the next value after '\n' (might be out of buffer bounds), 0 if '\n' not encountered yet
        **/
        nat_t push(char* buf, nat_t len) {
            nat_t i = 0;
            if (is_nan)
                goto NaN;
            for (; i < len; i++) {
                char c = buf[i];
                if (c == '\n')
                    return i + 1;
                if (c == '.' || c == ',') {
                    if (pt_pos != 0) // NaN
                        goto setNaN;
                    pt_pos = 1;
                    continue;
                }
                if (c < '0' || c > '9') // NaN
                    goto setNaN;
                if (pt_pos == 0) { // Before '.'
                    current = current * 10 + static_cast<val_t>(c - '0');
                } else { // After '.'
                    val_t n = static_cast<val_t>(c - '0');
                    for (nat_t j = 0; j < pt_pos; j++)
                        n /= 10;
                    current += n;
                    pt_pos++;
                }
            }
            return 0;
            NaN: {
                is_nan = true;
                for (; i < len; i++) {
                    char c = buf[i];
                    if (c == '\n')
                        return i + 1;
                }
                return 0;
            }
            setNaN: {
                i++;
                goto NaN;
            }
        }
        /** Reset the parser.
         * @return Value before reset
        **/
        val_t reset() {
            val_t ret = current;
            current = 0;
            pt_pos = 0;
            is_nan = false;
            return ret;
        }
    };
private:
    char* prog;     // Program path
    pid_t pid;      // Child process pid (0 for none)
    fd_t  pipes[2]; // Pipes used for communication
public:
    val_t latencies[3]; // Latencies for get, set, and remove operations respectively
public:
    /** Constructor.
     * @param prog Program path
    **/
    Test(char* prog): prog(prog), pid(0), pipes{ 0, 0 }, latencies{ 0 } {
    }
    /** Destructor.
    **/
    ~Test() {
        for (nat_t i = 0; i < 2; i++)
            if (pipes[i] != 0)
                close(pipes[i]);
    }
public:
    /** Run the tests with the given parameters.
     * @param cores Amount of cores to use
     * @param load  Percentage of modify operations
     * @param reps  Amount of repetitions (average)
     * @return True on success, false otherwise
    **/
    bool run(nat_t cores, nat_t load, nat_t reps = Constants::reps) {
        if (pipe(pipes) != 0) {
            std::cerr << "Unable to open pipes" << std::endl;
            return false;
        }
        pid = fork();
        if (pid == 0) { // Child
            Converter<> cores_text(cores);
            Converter<> update_text(load);
            Converter<> put_text(load / 2);
            char const* args[] = { prog, "-n", cores_text, "-u", update_text, "-p", put_text, "-o", null };
            char const* envs[] = { null };
            if (dup2(pipes[1], 1) == -1 || dup2(pipes[1], 2) == -1) {
                std::cerr << "Unable to set pipes" << std::endl;
                return false;
            }
            close(pipes[0]);
            execve(prog, const_cast<char**>(args), const_cast<char**>(envs));
            std::cerr << "Unable to start program" << std::endl;
            return false;
        }
        close(pipes[1]);
        return true;
    }
    /** Wait for child, while parsing its output.
    **/
    void wait() {
        constexpr nat_t size = 256;
        char buf[size];
        char* buffer;
        nat_t line = 0; // Current line (= current index in 'latencies')
        Parser parser;
        while (true) {
            buffer = buf;
            int len = read(pipes[0], buffer, size);
            if (len < 0) {
                std::cerr << "Unable to read pipe: " << len << std::endl;
                return;
            }
            if (len == 0)
                break;
        again:
            nat_t next = parser.push(buffer, len);
            if (next == 0) // Value not parsed entirely
                continue;
            latencies[line++] = parser.reset();
            if (line >= 3)
                break;
            if (next < static_cast<nat_t>(len)) { // 'len > 0' here
                buffer = buffer + next;
                len -= next;
                goto again;
            }
        }
        { // Wait for child to terminate
            int status; // Ignored
            waitpid(pid, &status, 0);
        }
    }
};

// ―――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――――

/** Program entry point.
 * @param argc Amount of arguments
 * @param argv Arguments
 * @return Return code
**/
int main(int argc, char** argv) {
    if (argc < 3) {
        std::cerr << "Usage: " << argv[0] << " <ldi name> <ldi binaries> ..." << std::endl;
        return 0;
    }
    std::string hostname;
    { // Init host name
        constexpr nat_t name_length = 16;
        char name[name_length + 1];
        name[name_length] = '\0';
        if (gethostname(name, name_length) != 0) {
            std::cerr << "Unable to get the host name" << std::endl;
            return 1;
        }
        hostname = name;
    }
    nat_t cores = sysconf(_SC_NPROCESSORS_ONLN) * 4 / 3; // Arbitrary, to see a natural increase in latencies
    nat_t core_divs = (Constants::core_divs >= cores ? cores - 1 : Constants::core_divs);
    char** bins = argv + 2;
    nat_t bins_length = argc - 2;
    for (nat_t load = 0; load < Constants::load_length; load++) { // Perform tests for each load
        nat_t load_perc = Constants::load[load];
        std::string filename;
        { // Build file name
            filename.append(hostname); // Host name
            filename.append(".");
            filename.append(argv[1]); // Group name
            filename.append(".u");
            { // Load value
                Converter<> load_text(load_perc);
                filename.append(load_text);
            }
            filename.append(".dat");
        }
        std::ofstream fout;
        { // Open and init file
            fout.open(filename);
            if (!fout.is_open()) {
                std::cerr << "Unable to write file '" << filename << "'" << std::endl;
                return 1;
            }
            fout << "#cores\t";
            for (nat_t bin = 0; bin < bins_length; bin++)
                fout << bins[bin] << "\t\t\t";
            fout << std::endl;
        }
        std::cout << "Output file '" << filename << "'" << std::endl;
        for (nat_t core = 0; core <= core_divs; core++) {
            nat_t cores_in_use = (core == 0 ? 1 : cores * core / core_divs); // Number of cores in use
            if (cores_in_use == 1 && core != 0)
                continue;
            std::cout << "  With " << cores_in_use << " core(s): ";
            fout << cores_in_use << "\t";
            for (nat_t bin = 0; bin < bins_length; bin++) {
                char* current_bin = bins[bin];
                if (bin != 0)
                    std::cout << ", ";
                std::cout << current_bin;
                std::cout.flush();
                val_t latencies[3] = { 0 };
                for (nat_t i = 0; i < Constants::reps; i++) {
                    Test test(current_bin);
                    if (!test.run(cores_in_use, load_perc)) // Unable to load the binary
                        return 1;
                    test.wait(); // Waiting results
                    for (nat_t j = 0; j < 3; j++)
                        latencies[j] += test.latencies[j] / Constants::reps;
                }
                for (nat_t i = 0; i < 3; i++)
                    fout << latencies[i] << "\t";
            }
            std::cout << std::endl;
            fout << std::endl;
        }
    }
    return 0;
}

// ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
// ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔ Test ▔
// ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔
