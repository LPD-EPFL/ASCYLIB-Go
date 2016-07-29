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
 * Channel measurement program.
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
    constexpr nat_t modes[] = { 0, 2 }; // Modes tested (random and shared)
    constexpr nat_t modes_length = sizeof(modes) / sizeof(nat_t); // Number of elements in table 'modes'
    constexpr char const* mode_names[] = { "random", "round-robin", "shared" }; // Mode names
    constexpr nat_t clients_divs = 16; // Number of divisions
    constexpr nat_t servers_divs = 4;  // Number of divisions
    static_assert(Constants::clients_divs >= 1, "Not enough divisions (at least 1)");
    static_assert(Constants::servers_divs >= 1, "Not enough divisions (at least 1)");
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
        val_t factor;  // Position factor, 1 for "before point"
        bool  is_nan;  // True if not a number, false otherwise
    public:
        /** Constructor.
        **/
        Parser(): current(0.), factor(1.), is_nan(false) {
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
                    if (factor != 1) // NaN
                        goto setNaN;
                    factor = 0.1;
                    continue;
                }
                if (c < '0' || c > '9') // NaN
                    goto setNaN;
                if (factor == 1) { // Before '.'
                    current = current * 10 + static_cast<val_t>(c - '0');
                } else { // After '.'
                    val_t n = static_cast<val_t>(c - '0');
                    current += factor * n;
                    factor /= 10;
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
            factor = 1;
            is_nan = false;
            return ret;
        }
    };
private:
    char* prog;     // Program path
    pid_t pid;      // Child process pid (0 for none)
    fd_t  pipes[2]; // Pipes used for communication
public:
    val_t values[3]; // Latencies for get, set, and remove operations respectively
public:
    /** Constructor.
     * @param prog Program path
    **/
    Test(char* prog): prog(prog), pid(0), pipes{ 0, 0 }, values{ 0 } {
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
     * @param mode    Mode to use
     * @param servers Amount of servers
     * @param clients Amount of clients
     * @return True on success, false otherwise
    **/
    bool run(nat_t mode, nat_t servers, nat_t clients) {
        if (pipe(pipes) != 0) {
            std::cerr << "Unable to open pipes" << std::endl;
            return false;
        }
        pid = fork();
        if (pid == 0) { // Child
            Converter<> mode_text(mode);
            Converter<> servers_text(servers);
            Converter<> clients_text(clients);
            char const* args[] = { prog, "-m", mode_text, "-s", servers_text, "-c", clients_text, "-o", null };
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
        nat_t line = 0; // Current line (= current index in 'values')
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
            values[line++] = parser.reset();
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
    if (argc != 2) {
        std::cerr << "Usage: " << argv[0] << " <channel binary>" << std::endl;
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
    nat_t servers = sysconf(_SC_NPROCESSORS_ONLN);
    nat_t servers_divs = (Constants::servers_divs >= servers ? servers : Constants::servers_divs);
    nat_t clients = servers * 2; // Arbitrary
    nat_t clients_divs = (Constants::clients_divs >= clients ? clients : Constants::clients_divs);
    for (nat_t mode_id = 0; mode_id < Constants::modes_length; mode_id++) { // Perform tests for each mode
        nat_t mode = Constants::modes[mode_id];
        for (nat_t server = 0; server <= servers_divs; server++) { // Perform tests for each server count
            nat_t servers_in_use = (server == 0 ? 1 : servers * server / servers_divs); // Number of servers in use
            if (servers_in_use == 1 && server != 0)
                continue;
            std::string filename;
            { // Build file name
                filename.append(hostname); // Host name
                filename.append(".");
                filename.append(Constants::mode_names[mode]); // Mode name
                filename.append(".s");
                filename.append(Converter<>(servers_in_use)); // Number of servers
                filename.append(".dat");
            }
            std::ofstream fout;
            { // Open and init file
                fout.open(filename);
                if (!fout.is_open()) {
                    std::cerr << "Unable to write file '" << filename << "'" << std::endl;
                    return 1;
                }
                fout << "#clients\t#messages\tthroughput (MB/s)\tlatency (µs)" << std::endl;
            }
            std::cout << "Output file '" << filename << "'" << std::endl;
            for (nat_t client = 0; client <= clients_divs; client++) {
                nat_t clients_in_use = (client == 0 ? 1 : clients * client / clients_divs); // Number of clients in use
                if (clients_in_use == 1 && client != 0)
                    continue;
                { // Print progression
                    Converter<> clients_text(clients_in_use);
                    std::cout << "  With " << clients_text << " client(s)... ";
                    std::cout.flush();
                    fout << clients_text << "\t";
                }
                { // Measurements
                    val_t values[3] = { 0. }; // Arithmetic means of { msg size (bytes), msg exchanges, test duration (ns) }
                    for (nat_t i = 0; i < Constants::reps; i++) { // Repetitions, keep average
                        Test test(argv[1]);
                        if (!test.run(mode, servers_in_use, clients_in_use)) // Unable to load the binary
                            return 1;
                        test.wait(); // Waiting results
                        for (nat_t j = 0; j < 3; j++)
                            values[j] += test.values[j] / Constants::reps;
                    }
                    { // Compute then output useful values
                        val_t throughput = values[1] * 1000. / values[2] * values[0]; // Global throughput (MB/s)
                        val_t latency = values[2] / 1000. / values[1] * static_cast<val_t>(clients_in_use); // Latency (µs) to send one message for one client
                        fout << values[1] << "\t" << throughput << "\t" << latency;
                    }
                }
                std::cout << "done." << std::endl;
                fout << std::endl;
            }
        }
    }
    return 0;
}

// ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
// ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔ Test ▔
// ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔
