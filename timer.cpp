#include <iostream>
#include <chrono>
#include <thread>

int main()
{
    int lineCount = 0;
    while (lineCount < 50)
    {
        char command;
        do
        {
            std::cin >> command;
        } while (command != 's');

        // Get the current time
        auto currentTime = std::chrono::system_clock::now();
        std::time_t currentTime_t = std::chrono::system_clock::to_time_t(currentTime);

        // Print the current time
        std::cout << "Current time: " << std::ctime(&currentTime_t) << std::flush;

        // Increment the line count
        lineCount++;

        // Sleep for 30 seconds
        std::this_thread::sleep_for(std::chrono::seconds(30));
    }

    return 0;
}
