import numpy as np
import matplotlib.pyplot as plt


BLOCKS_PER_YEAR = 2102400 

def __base_rate_per_block(base_rate):
    return base_rate/BLOCKS_PER_YEAR

def __multiplier_per_block(multiplier):
    return multiplier/BLOCKS_PER_YEAR

def __jump_multiplier_per_block(jump_multiplier):
    return jump_multiplier/BLOCKS_PER_YEAR

def __borrow_rate_per_block(utilization, base_rate, multiplier, jump_multiplier, kink):
    if kink == 0 or utilization <= kink:
        return utilization * __multiplier_per_block(multiplier) + __base_rate_per_block(base_rate)
    
    normal_rate = kink * __multiplier_per_block(multiplier) + __base_rate_per_block(base_rate)
    excess_util_rate = utilization - kink
    return excess_util_rate * __jump_multiplier_per_block(jump_multiplier) + normal_rate

def __supply_rate_per_block(utilization, base_rate, multiplier, jump_multiplier, kink, reserve_factor):
    borrow_rate = __borrow_rate_per_block(utilization, base_rate, multiplier, jump_multiplier, kink)
    one_minus_reserve_factor = 1 - reserve_factor
    rate_to_pool = borrow_rate * one_minus_reserve_factor
    return utilization * rate_to_pool

def __borrow_APY(utilization, base_rate, multiplier, jump_multiplier, kink):
    return __borrow_rate_per_block(utilization, base_rate, multiplier, jump_multiplier, kink) * BLOCKS_PER_YEAR

def __supply_APY(utilization, base_rate, multiplier, jump_multiplier, kink, reserve_factor):
    return __supply_rate_per_block(utilization, base_rate, multiplier, jump_multiplier, kink, reserve_factor) * BLOCKS_PER_YEAR

def supply_borrow_rate(base_rate, multiplier, jump_multiplier, kink, reserve_factor):
    x = np.linspace(0, 100, 100, endpoint=True)
    y_borrow_rate = []
    y_supply_rate = []

    print("utilizations:", x)

    for utilization in x:
        u = utilization / 100
        y_borrow_rate.append(__borrow_APY(u, base_rate, multiplier, jump_multiplier, kink))
        y_supply_rate.append(__supply_APY(u, base_rate, multiplier, jump_multiplier, kink, reserve_factor))
    
    print("borror_rates:", y_borrow_rate)
    print("supply_rates:", y_supply_rate)

    plt.figure(figsize=(8,4))
    plt.plot(x,y_borrow_rate, color="red",linewidth=2)
    plt.plot(x,y_supply_rate, color="green",linewidth=2)
    plt.xlabel("Utilization(%)")
    plt.ylabel("Rate(%)")
    plt.title("compound_rates")
    plt.legend()
    plt.show()
    plt.savefig("compound_rate.jpg")

if __name__ == "__main__":
    # supply_borrow_rate(0.0202, 0.2, 0.5, 0.8, 0.1)
    supply_borrow_rate(0, 0.058, 1.476, 0.8, 0.15)
