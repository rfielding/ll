#!/usr/bin/env python3
import random

# this is our function to fit with the data
# there are only three weights to deal with
def uslFunc(weights, n):
	alpha = weights[0] # contention w_0 is queing behind resources
	beta = weights[1] # coherence w_1 is causing others to do work for you
	gamma = weights[2] # processor speed, like in hertz for w_2
	return (gamma * n) / (1 + alpha*(n - 1) + beta * n * (n - 1))

def uslConstraints(weights,sample):
	if weights[0] < 0:
		weights[0] = 0
	if weights[0] >= 1.0:
		weights[0] = 1.0
	if weights[1] < 0:
		weights[1] = 0
	if weights[1] >= 1.0:
		weights[1] = 1.0
	if weights[2] <= 0:
		weights[2] = 0.001
	
# Tweak weights due to an observed sample
def Observe(F, constraints, weights, sample, step):
	constraints(weights,sample)
	# constants for this evaluation
	predicted = F(weights,sample[0])
	input = sample[0]
	actual = sample[1]
	derr = (actual-predicted)
	err = (0.5)*derr*derr 
	nextWeights = weights.copy()
	for i in range(0,len(nextWeights)):
		# tweak each weight input to find lowest err output weight
		w2 = weights.copy()
		w2[i] += step
		# grad (1/2)F^2 wrt w2_i
		nextWeights[i] += derr * (F(w2,sample[0])-predicted)
	for i in range(0,len(nextWeights)):
		weights[i] = nextWeights[i]
	#constraints(weights,sample)
	return err

observed = []

weightGuess = [0.01, 0.001, 1.0]
stepGuess = 0.00001

# create random data that fits the weight guess in some way
for k in range(1,10):
	for i in range(1,10):
		sample = [i, i*1.0]
		r = (1.0 + 0.02*(random.random() - 0.5))
		print("%f" % r)
		observed.append([
			i,
			r*uslFunc(weightGuess,i)],
		)
 
print("%s" % weightGuess)

for k in range(0, 10000):
	for i in range(0, len(observed)):
		err = Observe(uslFunc,uslConstraints,weightGuess,observed[i],stepGuess)
		#print("err: %f" % err)

print("(%f * n)/(1 + %f*(n-1) + %f*n*(n-1))" % (
	weightGuess[2],
	weightGuess[0],
	weightGuess[1],
))
